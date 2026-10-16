package config

import (
	"api-health-monitor/internal/model"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

func LoadEndpoints(filepath string) ([]model.Endpoint, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var endpoints []model.Endpoint

	if err := json.Unmarshal(file, &endpoints); err != nil {
		return nil, err
	}

	if err := validateEndpoints(endpoints); err != nil {
		return nil, err
	}

	return endpoints, nil
}

func validateEndpoint(endpoint model.Endpoint) error {
	if endpoint.ID <= 0 {
		return fmt.Errorf("endpoint %d: %w", endpoint.ID, model.ErrInvalidID)
	}

	raw := endpoint.URL
	if strings.TrimSpace(raw) == "" || raw != strings.TrimSpace(raw) {
		return fmt.Errorf("endpoint %d: %w", endpoint.ID, model.ErrInvalidURL)
	}

	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("endpoint %d: parse URL: %w", endpoint.ID, err)
	}

	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return fmt.Errorf("endpoint %d: %w", endpoint.ID, model.ErrInvalidURLScheme)
	}

	if u.Hostname() == "" {
		return fmt.Errorf("endpoint %d: %w", endpoint.ID, model.ErrMissingHostname)
	}

	return nil
}

func validateEndpoints(endpoints []model.Endpoint) error {
	if len(endpoints) == 0 {
		return model.ErrNoEndpoints
	}

	seen := make(map[int]struct{}, len(endpoints))
	for _, endpoint := range endpoints {
		if err := validateEndpoint(endpoint); err != nil {
			return err
		}
		
		if _, found := seen[endpoint.ID]; found {
			return fmt.Errorf("duplicate endpoint %d: %w", endpoint.ID, model.ErrDuplicateEndpoint)
		}

		seen[endpoint.ID] = struct{}{}
	}
	return nil
}