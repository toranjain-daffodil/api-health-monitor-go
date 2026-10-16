package model

import "time"

type Endpoint struct {
	ID  int    `json:"id" binding:"required"`
	URL string `json:"url" binding:"required"`
}

type EndpointResult struct {
	EndpointID          int
	URL                 string
	Status              string
	HTTPStatusCode      int
	Attempts            int
	LastAttemptDuration time.Duration
	TotalElapsed        time.Duration
	Error               string
}

type MonitorReport struct {
	Results          []EndpointResult
	SuccessCount     int
	FailureCount     int
	CancelledCount   int
	TotalRunDuration time.Duration
}
