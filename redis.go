package common

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

// NewRedisModule returns the redis health module.
func NewRedisModule(redis RedisClient, enabled bool) *RedisModule {
	return &RedisModule{
		redis:   redis,
		enabled: enabled,
	}
}

// RedisModule is the health check module for redis.
type RedisModule struct {
	redis   RedisClient
	enabled bool
}

// RedisClient is the interface of the redis client.
type RedisClient interface {
	Do(cmd string, args ...interface{}) (interface{}, error)
}

// RedisReport is the health report returned by the redis module.
type RedisReport struct {
	Name     string
	Duration time.Duration
	Status   Status
	Error    error
}

// MarshalJSON marshal the redis report.
func (r *RedisReport) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name     string `json:"name"`
		Duration string `json:"duration"`
		Status   string `json:"status"`
		Error    string `json:"error"`
	}{
		Name:     r.Name,
		Duration: r.Duration.String(),
		Status:   r.Status.String(),
		Error:    err(r.Error),
	})
}

// HealthChecks executes all health checks for Redis.
func (m *RedisModule) HealthChecks(context.Context) []RedisReport {
	if !m.enabled {
		return []RedisReport{{Name: "redis", Status: Deactivated}}
	}

	var reports = []RedisReport{}
	reports = append(reports, m.redisPingCheck())
	return reports
}

func (m *RedisModule) redisPingCheck() RedisReport {
	var healthCheckName = "ping"

	var now = time.Now()
	var _, err = m.redis.Do("PING")
	var duration = time.Since(now)

	var hcErr error
	var s Status
	switch {
	case err != nil:
		hcErr = errors.Wrap(err, "could not ping redis")
		s = KO
	default:
		s = OK
	}

	return RedisReport{
		Name:     healthCheckName,
		Duration: duration,
		Status:   s,
		Error:    hcErr,
	}
}

// MakeRedisModuleLoggingMW makes a logging middleware at module level.
func MakeRedisModuleLoggingMW(logger log.Logger) func(RedisHealthChecker) RedisHealthChecker {
	return func(next RedisHealthChecker) RedisHealthChecker {
		return &redisModuleLoggingMW{
			logger: logger,
			next:   next,
		}
	}
}

// RedisHealthChecker is the interface of the redis health check module.
type RedisHealthChecker interface {
	HealthChecks(context.Context) []RedisReport
}

// Logging middleware at module level.
type redisModuleLoggingMW struct {
	logger log.Logger
	next   RedisHealthChecker
}

// redisModuleLoggingMW implements RedisHealthChecker. There must be a key "correlation_id" with a string value in the context.
func (m *redisModuleLoggingMW) HealthChecks(ctx context.Context) []RedisReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}
