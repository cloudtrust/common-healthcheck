package common

import (
	"context"
	"net/http"
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

const (
	agentSystemDUnitName = "agent.service"
)

// JaegerModule is the health check module for jaeger.
type JaegerModule struct {
	conn                    SystemDConn
	collectorHealthCheckURL string
	httpClient              JaegerHTTPClient
	enabled                 bool
}

// systemDConn is interface of systemd D-Bus connection.
type SystemDConn interface {
	ListUnitsByNames(units []string) ([]dbus.UnitStatus, error)
}

// jaegerHTTPClient is the interface of the http client.
type JaegerHTTPClient interface {
	Get(string) (*http.Response, error)
}

// NewJaegerModule returns the jaeger health module.
func NewJaegerModule(conn SystemDConn, httpClient JaegerHTTPClient, collectorHealthCheckURL string, enabled bool) *JaegerModule {
	return &JaegerModule{
		conn:                    conn,
		httpClient:              httpClient,
		collectorHealthCheckURL: collectorHealthCheckURL,
		enabled:                 enabled,
	}
}

// JaegerReport is the health report returned by the jaeger module.
type JaegerReport struct {
	Name     string
	Duration time.Duration
	Status   Status
	Error    error
}

// HealthChecks executes all health checks for Jaeger.
func (m *JaegerModule) HealthChecks(context.Context) []JaegerReport {
	var reports = []JaegerReport{}
	reports = append(reports, m.jaegerSystemDCheck())
	reports = append(reports, m.jaegerCollectorPing())
	return reports
}

func (m *JaegerModule) jaegerSystemDCheck() JaegerReport {
	var healthCheckName = "jaeger agent systemd unit check"

	if !m.enabled {
		return JaegerReport{
			Name:   healthCheckName,
			Status: Deactivated,
		}
	}

	var now = time.Now()
	var units, err = m.conn.ListUnitsByNames([]string{agentSystemDUnitName})
	var duration = time.Since(now)

	var hcErr error
	var s Status
	switch {
	case err != nil:
		hcErr = errors.Wrapf(err, "could not list '%s' systemd unit", agentSystemDUnitName)
		s = KO
	case len(units) == 0:
		hcErr = errors.Wrapf(err, "systemd unit '%s' not found", agentSystemDUnitName)
		s = KO
	case units[0].ActiveState != "active":
		hcErr = errors.Wrapf(err, "systemd unit '%s' is not active", agentSystemDUnitName)
		s = KO
	default:
		s = OK
	}

	return JaegerReport{
		Name:     healthCheckName,
		Duration: duration,
		Status:   s,
		Error:    hcErr,
	}
}

func (m *JaegerModule) jaegerCollectorPing() JaegerReport {
	var healthCheckName = "ping jaeger collector"

	if !m.enabled {
		return JaegerReport{
			Name:   healthCheckName,
			Status: Deactivated,
		}
	}

	// query jaeger collector health check URL
	var now = time.Now()
	var res, err = m.httpClient.Get("http://" + m.collectorHealthCheckURL)
	var duration = time.Since(now)

	var hcErr error
	var s Status
	switch {
	case err != nil:
		hcErr = errors.Wrap(err, "could not query jaeger collector health check service")
		s = KO
	case res.StatusCode != 204:
		hcErr = errors.Wrapf(err, "jaeger health check service returned invalid status code: %v", res.StatusCode)
		s = KO
	default:
		s = OK
	}

	return JaegerReport{
		Name:     healthCheckName,
		Duration: duration,
		Status:   s,
		Error:    hcErr,
	}
}

// JaegerHealthChecker is the interface of the jaeger health check module.
type JaegerHealthChecker interface {
	HealthChecks(context.Context) []JaegerReport
}

// Logging middleware at module level.
type jaegerModuleLoggingMW struct {
	logger log.Logger
	next   JaegerHealthChecker
}

// MakeJaegerModuleLoggingMW makes a logging middleware at module level.
func MakeJaegerModuleLoggingMW(logger log.Logger) func(JaegerHealthChecker) JaegerHealthChecker {
	return func(next JaegerHealthChecker) JaegerHealthChecker {
		return &jaegerModuleLoggingMW{
			logger: logger,
			next:   next,
		}
	}
}

// jaegerModuleLoggingMW implements Module.
func (m *jaegerModuleLoggingMW) HealthChecks(ctx context.Context) []JaegerReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}
