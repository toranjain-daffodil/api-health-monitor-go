package model

import "errors"

var (
	ErrInvalidID = errors.New("invalid endpoint id")
	ErrInvalidURL = errors.New("invalid endpoint url")
	ErrInvalidURLScheme = errors.New("url scheme must be http or https")
	ErrMissingHostname = errors.New("hostname is required")
	ErrNoEndpoints = errors.New("endpoints are required")
	ErrDuplicateEndpoint = errors.New("endpoint id must be unique")
)