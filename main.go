package main

import (
	"api-health-monitor/internal/checker"
	"api-health-monitor/internal/config"
	"api-health-monitor/internal/model"
	"api-health-monitor/internal/monitor"
	"context"
	"log"
	"net/http"
	"time"
)

func main() {
	endpoints, err := config.LoadEndpoints("./config/endpoints.json")
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	checker := checker.NewChecker(&client)
	cfg := model.MonitorConfig {
		MaxWorkers: 10,
		MaxRetries: 2,
		RetryBackoff: 200 * time.Millisecond,
	}
	monitor := monitor.NewMonitor(checker, &cfg)
	res, err := monitor.Start(context.Background(), endpoints)
	if err != nil {
		log.Fatal(err)
	}
	
	for _, r:= range res.Results {
		log.Println("=================================")
		log.Println("ID:", r.EndpointID)
		log.Println("URL:", r.URL)
		log.Println("Status:", r.Status)
		log.Println("Status Code:", r.HTTPStatusCode)
		log.Println("Last Attempt Duration:", r.LastAttemptDuration)
		log.Println("Total Elapsed Time:", r.TotalElapsed)
		log.Println("Error:", r.Error)
		log.Println("=================================")
	}
	log.Println("=================================")
	log.Println("RESULTS")
	log.Println("Success Count:", res.SuccessCount)
	log.Println("Cancellation Count:", res.CancelledCount)
	log.Println("Failure Count", res.FailureCount)
	log.Println("Total Run Duration:", res.TotalRunDuration)
	log.Println("=================================")
}
