
package checker

import (
	"api-health-monitor/internal/model"
	"context"
	"fmt"
	"net/http"
	"time"
)

type Checker interface {
	Check(ctx context.Context, endpoint model.Endpoint) model.EndpointResult
}

type checker struct {
	client *http.Client
}

func NewChecker(client *http.Client) *checker {
	if client == nil {
		client = &http.Client{
			Timeout: 3 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	return &checker{
		client: client,
	}
}

func (c *checker) Check(ctx context.Context, endpoint model.Endpoint) model.EndpointResult {
	start := time.Now()

	result := model.EndpointResult{
		EndpointID: endpoint.ID,
		URL:        endpoint.URL,
	}

	finishCancelled := func(err error) model.EndpointResult {
		result.Status = "cancelled"
		result.Error = fmt.Sprintf("context ended: %v", err)
		result.LastAttemptDuration = time.Since(start)
		return result
	}

	if err := ctx.Err(); err != nil {
		return finishCancelled(err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint.URL,
		nil,
	)
	if err != nil {
		result.Status = "failure"
		result.Error = fmt.Sprintf("create request: %v", err)
		result.LastAttemptDuration = time.Since(start)
		return result
	}

	resp, err := c.client.Do(req)
	result.LastAttemptDuration = time.Since(start)

	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return finishCancelled(ctxErr)
		}

		result.Status = "failure"
		result.Error = fmt.Sprintf("HTTP request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.HTTPStatusCode = resp.StatusCode

	if resp.StatusCode >= http.StatusOK &&
		resp.StatusCode < http.StatusMultipleChoices {
		result.Status = "success"
		return result
	}

	result.Status = "failure"
	result.Error = fmt.Sprintf("unexpected HTTP status: %d", resp.StatusCode)

	return result
}
