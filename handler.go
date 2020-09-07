package main

import (
	"encoding/json"
	"fmt"
	"github.com/mercadolibre/golang-restclient/rest"
	"log"
	"net/http"
	"time"
)

var (
	baseUrlChildren = "https://api.mercadolibre.com/items/%s/children/"

	baseUrlItem = "https://api.mercadolibre.com/items/%s/"
)

func ItemHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var transaction [2]*rest.FutureResponse

	var Children []Child

	var item Item

	//ejecutar los request concurrentemente y orquestando las llamadas
	rest.ForkJoin(func(c *rest.Concurrent) {
		transaction[0] = c.Get(fmt.Sprintf(baseUrlChildren, r.RequestURI))
		transaction[1] = c.Get(fmt.Sprintf(baseUrlItem, r.RequestURI))
	})

	for _, tx := range transaction {
		log.Printf("tx is null: %v", tx == nil)
		log.Printf("response is null: %v", tx.Response() == nil)
		log.Printf("StatusCode is: %v", tx.Response().StatusCode)
		if tx.Response().StatusCode != http.StatusOK {
			ServerResponse(w, string(tx.Response().Bytes()), tx.Response().StatusCode)
			return
		}
	}
	//Unmarshal respuesta
	if err := json.Unmarshal(transaction[0].Response().Bytes(), &Children); err != nil {
		log.Println(err)

		ServerResponse(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(transaction[1].Response().Bytes(), &item); err != nil {
		log.Println(err)

		ServerResponse(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	item.Children = Children

	//Marshal respuesta
	marshal, err := json.Marshal(item)
	if err != nil {
		log.Println(err)
		ServerResponse(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}
	//Asignar cache de forma async
	go item.SetCache()

	//tracking logging health
	if len(HealthLogList) > 0 {
		lh := &HealthLogList[len(HealthLogList)-1]
		lh.TotalCountApiCall++
		lh.AvgResponseTimeApiCalls = (lh.AvgResponseTimeApiCalls + time.Since(start).Seconds()) / lh.TotalCountApiCall
		log.Printf("Health object  TotalCountApiCall %f AvgResponseTimeApiCalls %f\n", lh.TotalCountApiCall, lh.AvgResponseTimeApiCalls)
	}

	ServerResponse(w, string(marshal), http.StatusOK)
	log.Println("ItemHandler response")

	return
}

func HealthCheckerHandler(w http.ResponseWriter, r *http.Request) {
	marshal, err := json.Marshal(HealthLogList)
	if err != nil {
		log.Println(err)
		ServerResponse(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}
	ServerResponse(w, string(marshal), http.StatusOK)
	return
}
