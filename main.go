package main

import (
	"api-health-monitor/internal/checker"
	"api-health-monitor/internal/config"
	"api-health-monitor/internal/model"
	"api-health-monitor/internal/monitor"
	"api-health-monitor/internal/report"
	"context"
	"log"
	"net/http"
	"time"
)

func main() {
	endpoints, err := config.LoadEndpoints("./assets/endpoints.json")
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
	
	report.Display(res)
	if err := report.Save(res, "./assets/reports"); err != nil {
		log.Println(err)
	}
}
