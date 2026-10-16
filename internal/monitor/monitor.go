package monitor

import (
	"api-health-monitor/internal/checker"
	"api-health-monitor/internal/model"
	"context"
	"fmt"
	"sync"
	"time"
)

type Monitor struct {
	checker checker.Checker
	workerCount int
	maxRetries int
	retryBackoff time.Duration
}

func NewMonitor(checker checker.Checker, config *model.MonitorConfig) *Monitor {
	var cfg model.MonitorConfig
	if config != nil {
		cfg = *config
	}

	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = 10
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 2
	}
	if cfg.RetryBackoff == 0 {
		cfg.RetryBackoff = 200 * time.Millisecond
	}

	return &Monitor{
		checker: checker,
		workerCount: cfg.MaxWorkers,
		maxRetries: cfg.MaxRetries,
		retryBackoff: cfg.RetryBackoff,
	}
}

func (m Monitor) Start(c context.Context, endpoints []model.Endpoint) (model.MonitorReport, error) {
	start := time.Now()
	jobs := make(chan model.Endpoint)
	results := make(chan model.EndpointResult, len(endpoints))

	var wg sync.WaitGroup

	go func(){
		defer close(jobs)
		for _, endpoint := range endpoints {
			select {
			case <-c.Done():
				return
			case jobs <- endpoint:
			}
		}
	}()

	for range m.workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for e := range jobs {
				res := m.retryHelper(c, e)
				results <- res
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	resByID := make(map[int]model.EndpointResult, len(endpoints))
	var collectionErr error

	for result := range results {
		switch result.Status {
		case "success", "failure", "cancelled":
			resByID[result.EndpointID] = result
		default:
			if collectionErr == nil {
				collectionErr = fmt.Errorf("unexpected status received: %q", result.Status)
			}
		}
	}

	if collectionErr != nil {
		return model.MonitorReport{}, collectionErr
	}

	res := make([]model.EndpointResult, 0, len(endpoints))

	for _, endpoint := range endpoints {
		result, found := resByID[endpoint.ID]

		if !found {
			if c.Err() == nil {
				return model.MonitorReport{}, fmt.Errorf("missing result for endpoint %d", endpoint.ID)
			}
			result = model.EndpointResult{
				EndpointID: endpoint.ID,
				URL: endpoint.URL,
				Status: "cancelled",
				Error: fmt.Sprintf("context ended: %v", c.Err()),
			}
		}
		res = append(res, result)
	}

	success := 0
	failure := 0
	cancelled := 0

	for _, r := range res {
		switch r.Status {
		case "success":
			success += 1
		case "failure":
			failure += 1
		case "cancelled":
			cancelled += 1
		}
	}

	dur := time.Since(start)

	return model.MonitorReport{
		Results: res,
		SuccessCount: success,
		FailureCount: failure,
		CancelledCount: cancelled,
		TotalRunDuration: dur,
	}, nil
}

func (m Monitor) retryHelper(
	c context.Context,
	endpoint model.Endpoint,
) model.EndpointResult {
	start := time.Now()

	res := model.EndpointResult{
		EndpointID: endpoint.ID,
		URL:        endpoint.URL,
	}

	for attempt := 1; attempt <= m.maxRetries+1; attempt++ {
		if err := c.Err(); err != nil {
			res.Status = "cancelled"
			res.Error = fmt.Sprintf("context ended: %v", err)
			res.TotalElapsed = time.Since(start)
			return res
		}

		res = m.checker.Check(c, endpoint)
		res.Attempts = attempt
		res.TotalElapsed = time.Since(start)

		if err := c.Err(); err != nil || res.Status == "cancelled" {
			res.Status = "cancelled"

			if err != nil {
				res.Error = fmt.Sprintf("context ended: %v", err)
			}

			res.TotalElapsed = time.Since(start)
			return res
		}

		if res.Status == "success" {
			return res
		}

		if attempt == m.maxRetries+1 {
			return res
		}

		timer := time.NewTimer(m.retryBackoff)

		select {
		case <-c.Done():
			timer.Stop()
			res.Status = "cancelled"
			res.Error = fmt.Sprintf("context ended: %v", c.Err())
			res.TotalElapsed = time.Since(start)
			return res

		case <-timer.C:
		}
	}

	return res
}