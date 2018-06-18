package common

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

// NewFlakiModule returns the Flaki health module.
func NewFlakiModule(client FlakiClient) *FlakiModule {
	return &FlakiModule{
		flakiClient: client,
	}
}

// FlakiModule is the health check module for Flaki.
type FlakiModule struct {
	flakiClient FlakiClient
}

// FlakiClient is the interface of Flaki.
type FlakiClient interface {
	NextValidID() (string, error)
}

// FlakiReport is the health report returned by the flaki module.
type FlakiReport struct {
	Name     string
	Duration time.Duration
	Status   Status
	Error    error
}

// MarshalJSON marshal the flaki report.
func (r *FlakiReport) MarshalJSON() ([]byte, error) {
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

// HealthChecks executes all health checks for Flaki.
func (m *FlakiModule) HealthChecks(context.Context) []FlakiReport {
	var reports = []FlakiReport{}
	reports = append(reports, m.flakiNextIDCheck())
	return reports
}

func (m *FlakiModule) flakiNextIDCheck() FlakiReport {
	var healthCheckName = "Flaki ID generation"

	// query flaki next valid ID
	var now = time.Now()
	var _, err = m.flakiClient.NextValidID()
	var duration = time.Since(now)

	var hcErr error
	var s Status
	switch {
	case err != nil:
		hcErr = errors.Wrap(err, "could not query flaki service")
		s = KO
	default:
		s = OK
	}

	return FlakiReport{
		Name:     healthCheckName,
		Duration: duration,
		Status:   s,
		Error:    hcErr,
	}
}

// MakeFlakiModuleLoggingMW makes a logging middleware at module level.
func MakeFlakiModuleLoggingMW(logger log.Logger) func(FlakiHealthChecker) FlakiHealthChecker {
	return func(next FlakiHealthChecker) FlakiHealthChecker {
		return &flakiModuleLoggingMW{
			logger: logger,
			next:   next,
		}
	}
}

// FlakiHealthChecker is the interface of the flaki health check module.
type FlakiHealthChecker interface {
	HealthChecks(context.Context) []FlakiReport
}

// Logging middleware at module level.
type flakiModuleLoggingMW struct {
	logger log.Logger
	next   FlakiHealthChecker
}

// flakiModuleLoggingMW implements FlakiHealthChecker. There must be a key "correlation_id" with a string value in the context.
func (m *flakiModuleLoggingMW) HealthChecks(ctx context.Context) []FlakiReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}
