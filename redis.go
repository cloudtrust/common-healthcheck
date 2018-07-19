package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// RedisCheckNames contains the list of all valid tests names.
var RedisCheckNames = map[string]struct{}{
	"":     struct{}{},
	"ping": struct{}{},
}

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

type redisReport struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

// HealthCheck executes the desired influx health check.
func (m *RedisModule) HealthCheck(_ context.Context, name string) (json.RawMessage, error) {
	if !m.enabled {
		return json.MarshalIndent([]influxReport{{Name: "redis", Status: Deactivated.String()}}, "", "  ")
	}

	var reports []redisReport
	switch name {
	case "":
		reports = append(reports, m.redisPing())
	case "ping":
		reports = append(reports, m.redisPing())
	default:
		// Should not happen: there is a middleware validating the inputs name.
		panic(fmt.Sprintf("Unknown redis health check name: %v", name))
	}

	return json.MarshalIndent(reports, "", "  ")
}

func (m *RedisModule) redisPing() redisReport {
	var name = "ping"
	var status = OK

	var now = time.Now()
	var _, err = m.redis.Do("PING")
	var duration = time.Since(now)

	if err != nil {
		status = KO
		err = errors.Wrap(err, "could not ping redis")
	}

	return redisReport{
		Name:     name,
		Duration: duration.String(),
		Status:   status.String(),
		Error:    str(err),
	}
}
