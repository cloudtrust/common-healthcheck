package common

import (
	"context"
	"encoding/json"
)

// MakeValidationMiddleware makes a middleware that validate the health check name comming from
// the HTTP route.
func MakeValidationMiddleware(validValues map[string]struct{}) func(HealthChecker) HealthChecker {
	return func(next HealthChecker) HealthChecker {
		return &validationMW{
			validValues: validValues,
			next:        next,
		}
	}
}

type validationMW struct {
	validValues map[string]struct{}
	next        HealthChecker
}

type ErrInvalidHCName struct {
	s string
}

func (e *ErrInvalidHCName) Error() string {
	return e.s
}

func (m *validationMW) HealthCheck(ctx context.Context, name string) (json.RawMessage, error) {
	// Check health check name validity.
	var _, ok = m.validValues[name]
	if !ok {
		return nil, &ErrInvalidHCName{}
	}

	return m.next.HealthCheck(ctx, name)
}
