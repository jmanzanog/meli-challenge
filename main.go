package main

import (
	"github.com/robfig/cron"
	"log"
	"time"
)

var HealthLogList []Health

func main() {
	log.Print("start app")
	///init Logging Health
	HealthLogList = append(HealthLogList, Health{
		Date:                    time.Now(),
		AvgResponseTime:         0,
		TotalRequest:            0,
		AvgResponseTimeApiCalls: 0,
		TotalCountApiCall:       0,
	})

	c := cron.New()
	//init login cron
	go func() {
		err := c.AddFunc("*/60 * * * *", func() { addHealthLog() })
		if err != nil {
			log.Panicf("Error Adding cron %v", err)
		}
		c.Start()

	}()

	// init http server
	server := NewServer(ServerPortDefaultValue)

	//item handler + CacheMiddleware
	server.Handle("/items/", GetMethod, server.AddMiddleware(ItemHandler, CacheMiddleware(), LoggingMiddleware()))

	server.Handle("/health", GetMethod, server.AddMiddleware(HealthCheckerHandler, LoggingMiddleware()))

	// start server
	err := server.Start()
	if err != nil {
		log.Panicln(UnableServerMsg, err)
		return
	}
	log.Print("server running")
}

func addHealthLog() {
	log.Printf("HealthLogList size: %d\n", len(HealthLogList))
	HealthLogList = append(HealthLogList, Health{
		Date:                    time.Now(),
		AvgResponseTime:         0,
		TotalRequest:            0,
		AvgResponseTimeApiCalls: 0,
		TotalCountApiCall:       0,
	})
	log.Print("[Job] Add Log entry every minute job\n")
}
