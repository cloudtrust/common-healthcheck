package common

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kit/kit/log"
)

// MakeHealthCheckerLoggingMW makes a logging middleware for the health check modules.
func MakeHealthCheckerLoggingMW(logger log.Logger) func(HealthChecker) HealthChecker {
	return func(next HealthChecker) HealthChecker {
		return &healthCheckerLoggingMW{
			logger: logger,
			next:   next,
		}
	}
}

type healthCheckerLoggingMW struct {
	logger log.Logger
	next   HealthChecker
}

// healthCheckLoggingMW implements HealthChecker. There must be a key "correlation_id" with a string value in the context.
func (m *healthCheckerLoggingMW) HealthCheck(ctx context.Context, name string) (json.RawMessage, error) {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthCheck", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthCheck(ctx, name)
}
