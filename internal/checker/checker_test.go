package checker

import (
	"api-health-monitor/internal/model"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheck_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	checker := NewChecker(&client)

	endpoint := model.Endpoint{
		ID:  1,
		URL: server.URL,
	}

	result := checker.Check(context.Background(), endpoint)

	if result.Status != "success" {
		t.Fatalf("expected status success, got %q; error: %s", result.Status, result.Error)
	}

	if result.HTTPStatusCode != http.StatusOK {
		t.Errorf("expected HTTP status %d, got %d", http.StatusOK, result.HTTPStatusCode)
	}

	if result.LastAttemptDuration < 0 {
		t.Errorf("expected non-negative duration, got %v", result.LastAttemptDuration)
	}
}

func TestCheck_ServiceUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	checker := NewChecker(&client)

	endpoint := model.Endpoint{
		ID:  2,
		URL: server.URL,
	}

	result := checker.Check(context.Background(), endpoint)

	if result.Status != "failure" {
		t.Fatalf("expected status failure, got %q", result.Status)
	}

	if result.HTTPStatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected HTTP status %d, got %d",
			http.StatusServiceUnavailable, result.HTTPStatusCode)
	}

	if result.Error == "" {
		t.Error("expected an error message for unsuccessful HTTP status")
	}
}

func TestCheck_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(4 * time.Second):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			return
		}
	}))
	defer server.Close()

	client := http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	checker := NewChecker(&client)

	endpoint := model.Endpoint{
		ID:  3,
		URL: server.URL,
	}

	result := checker.Check(context.Background(), endpoint)

	if result.Status != "failure" {
		t.Fatalf("expected status failure, got %q", result.Status)
	}

	if result.Error == "" {
		t.Error("expected an error message for timeout")
	}

	if result.LastAttemptDuration < 3*time.Second {
		t.Errorf("expected duration to be at least 3 seconds, got %v",
			result.LastAttemptDuration)
	}
}
