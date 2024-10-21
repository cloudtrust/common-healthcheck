package common

import (
	"context"
	"encoding/json"
	"net/http"
)

// Status is the status of the health check.
type Status int

const (
	// OK is the status for a successful health check.
	OK Status = iota
	// KO is the status for an unsuccessful health check.
	KO
	// Deactivated is the status for a service that is deactivated, e.g. we can disable error tracking, instrumenting, tracing,...
	Deactivated
)

// HealthChecker is the interface of the health check modules.
type HealthChecker interface {
	HealthCheck(context.Context, string) (json.RawMessage, error)
}

// HTTPClient is the interface of the http client used to get health check status.
type HTTPClient interface {
	Get(string) (*http.Response, error)
}

// str return the string error that will be in the health report
func str(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
