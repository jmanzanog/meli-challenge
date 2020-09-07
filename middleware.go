package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

//LoggingMiddleware logguea los tiempos del handlers que se
func LoggingMiddleware() Middleware {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			start := time.Now()
			defer func() {
				if len(HealthLogList) > 0 {
					lh := &HealthLogList[len(HealthLogList)-1]
					lh.TotalRequest++
					lh.AvgResponseTime = (lh.AvgResponseTime + time.Since(start).Seconds()) / lh.TotalRequest
					log.Printf("LoggingMiddleware Logging list size: %d", len(HealthLogList))
					log.Printf("Health object  AvgResponseTime %f TotalRequest %f\n", lh.AvgResponseTime, lh.AvgResponseTime)
				}
			}()
			handlerFunc(writer, request)
		}
	}
}

//CacheMiddleware Cache de respuesta para items
func CacheMiddleware() Middleware {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Println("Start CacheMiddleware")
			//validar path
			uriList := strings.Split(r.RequestURI, "/")
			if len(uriList) > 3 {
				ServerResponse(w, fmt.Sprintf("path not found %s", uriList[3]), http.StatusNotFound)
				return
			}
			item := Item{}
			item.ItemID = uriList[2]

			//obteniedo cache de item
			err, itemCache := item.GetCache()
			if err != nil {
				log.Printf("no cache for %s reason: %v", item.ItemID, err)
				r.RequestURI = item.ItemID
				handlerFunc(w, r)
				return
			}
			// Marshal respuesta
			item = *itemCache
			marshal, _ := json.Marshal(item)
			ServerResponse(w, string(marshal), http.StatusOK)
			log.Println("cache response")
			return

		}
	}
}
