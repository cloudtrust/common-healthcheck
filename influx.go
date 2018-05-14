package common

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

// InfluxModule is the health check module for influx.
type InfluxModule struct {
	influx  Influx
	enabled bool
}

// influx is the interface of the influx client.
type Influx interface {
	Ping(timeout time.Duration) (time.Duration, string, error)
}

// NewInfluxModule returns the influx health module.
func NewInfluxModule(influx Influx, enabled bool) *InfluxModule {
	return &InfluxModule{
		influx:  influx,
		enabled: enabled,
	}
}

func (m *InfluxModule) influxPing() InfluxReport {
	var healthCheckName = "ping"

	if !m.enabled {
		return InfluxReport{
			Name:   healthCheckName,
			Status: Deactivated,
		}
	}

	var now = time.Now()
	var _, _, err = m.influx.Ping(5 * time.Second)
	var duration = time.Since(now)

	var hcErr error
	var s Status
	switch {
	case err != nil:
		hcErr = errors.Wrap(err, "could not ping influx")
		s = KO
	default:
		s = OK
	}

	return InfluxReport{
		Name:     healthCheckName,
		Duration: duration,
		Status:   s,
		Error:    hcErr,
	}
}

// InfluxReport is the health report returned by the influx module.
type InfluxReport struct {
	Name     string
	Duration time.Duration
	Status   Status
	Error    error
}

func (i *InfluxReport) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name     string `json:"name"`
		Duration string `json:"duration"`
		Status   string `json:"status"`
		Error    string `json:"error"`
	}{
		Name: i.Name,
		Duration: i.Duration.String(),
		Status: i.Status.String(),
		Error: err(i.Error),
	})
}

// HealthChecks executes all health checks for influx.
func (m *InfluxModule) HealthChecks(context.Context) []InfluxReport {
	var reports = []InfluxReport{}
	reports = append(reports, m.influxPing())
	return reports
}

// InfluxHealthChecker is the interface of the influx health check module.
type InfluxHealthChecker interface {
	HealthChecks(context.Context) []InfluxReport
}

// influxModuleLoggingMW implements Module.
func (m *influxModuleLoggingMW) HealthChecks(ctx context.Context) []InfluxReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}

// Logging middleware at module level.
type influxModuleLoggingMW struct {
	logger log.Logger
	next   InfluxHealthChecker
}

// MakeInfluxModuleLoggingMW makes a logging middleware at module level.
func MakeInfluxModuleLoggingMW(logger log.Logger) func(InfluxHealthChecker) InfluxHealthChecker {
	return func(next InfluxHealthChecker) InfluxHealthChecker {
		return &influxModuleLoggingMW{
			logger: logger,
			next:   next,
		}
	}
}
