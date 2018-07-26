package common

import (
	"context"
	"encoding/json"
	"fmt"
)

// MakeValidationMiddleware makes a middleware that validate the health check name comming from
// the HTTP route.
// Let's imagine we have a module with the healthchecks "ping", "create", and "delete".
// In the validValues map, we list the health check authorized, with "" that is a special value
// for executing all healthchecks.
//
// var validValues = map[string]struct{}{
//   "":       struct{}{},
// 	 "ping":   struct{}{},
//   "create": struct{}{},
//   "delete": struct{}{},
// }
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

// ErrInvalidHCName is the error returned when there is a health request for
// an unknown healthcheck name.
type ErrInvalidHCName struct {
	s string
}

func (e *ErrInvalidHCName) Error() string {
	return fmt.Sprintf("no health check with name '%s'", e.s)
}

func (m *validationMW) HealthCheck(ctx context.Context, name string) (json.RawMessage, error) {
	// Check health check name validity.
	var _, ok = m.validValues[name]
	if !ok {
		return nil, &ErrInvalidHCName{name}
	}

	return m.next.HealthCheck(ctx, name)
}
