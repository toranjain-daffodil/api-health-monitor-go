package model

import (
	"time"
)

type MonitorConfig struct {
	MaxWorkers int
	MaxRetries int
	RetryBackoff time.Duration
}