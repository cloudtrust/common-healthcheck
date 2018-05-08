package common

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/go-kit/kit/log"
)

// RedisModule is the health check module for redis.
type RedisModule struct {
	redis   RedisClient
	enabled bool
}

// redisClient is the interface of the redis client.
type RedisClient interface {
	Do(cmd string, args ...interface{}) (interface{}, error)
}

// NewRedisModule returns the redis health module.
func NewRedisModule(redis RedisClient, enabled bool) *RedisModule {
	return &RedisModule{
		redis:   redis,
		enabled: enabled,
	}
}

// RedisReport is the health report returned by the redis module.
type RedisReport struct {
	Name     string
	Duration time.Duration
	Status   Status
	Error    error
}

// HealthChecks executes all health checks for Redis.
func (m *RedisModule) HealthChecks(context.Context) []RedisReport {
	var reports = []RedisReport{}
	reports = append(reports, m.redisPingCheck())
	return reports
}

func (m *RedisModule) redisPingCheck() RedisReport {
	var healthCheckName = "ping"

	if !m.enabled {
		return RedisReport{
			Name:   healthCheckName,
			Status: Deactivated,
		}
	}

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

// RedisHealthChecker is the interface of the redis health check module.
type RedisHealthChecker interface {
	HealthChecks(context.Context) []RedisReport
}




// Logging middleware at module level.
type redisModuleLoggingMW struct {
	logger log.Logger
	next   RedisHealthChecker
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

// redisModuleLoggingMW implements Module.
func (m *redisModuleLoggingMW) HealthChecks(ctx context.Context) []RedisReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}