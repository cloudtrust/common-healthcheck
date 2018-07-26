package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/pkg/errors"
)

const (
	agentSystemDUnitName = "agent.service"
)

// NewJaegerModule returns the jaeger health module.
func NewJaegerModule(conn SystemDConn, httpClient HTTPClient, collectorHealthHostPort string, enabled bool) *JaegerModule {
	return &JaegerModule{
		conn:                    conn,
		httpClient:              httpClient,
		collectorHealthHostPort: collectorHealthHostPort,
		enabled:                 enabled,
	}
}

// JaegerModule is the health check module for jaeger.
type JaegerModule struct {
	conn                    SystemDConn
	collectorHealthHostPort string
	httpClient              HTTPClient
	enabled                 bool
}

// SystemDConn is interface of systemd D-Bus connection.
type SystemDConn interface {
	ListUnitsByNames(units []string) ([]dbus.UnitStatus, error)
}

type jaegerReport struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

// HealthCheck executes the desired jaeger health check.
func (m *JaegerModule) HealthCheck(_ context.Context, name string) (json.RawMessage, error) {
	if !m.enabled {
		return json.MarshalIndent([]jaegerReport{{Name: "jaeger", Status: Deactivated.String()}}, "", "  ")
	}

	var reports []jaegerReport
	switch name {
	case "":
		reports = append(reports, m.jaegerSystemDCheck())
		reports = append(reports, m.jaegerCollectorPing())
	case "agent":
		reports = append(reports, m.jaegerSystemDCheck())
	case "collector":
		reports = append(reports, m.jaegerCollectorPing())
	default:
		// Should not happen: there is a middleware validating the inputs name.
		panic(fmt.Sprintf("Unknown jaeger health check name: %v", name))
	}

	return json.MarshalIndent(reports, "", "  ")
}

func (m *JaegerModule) jaegerSystemDCheck() jaegerReport {
	var name = "agent systemd unit"
	var status = OK

	var now = time.Now()
	var units, err = m.conn.ListUnitsByNames([]string{agentSystemDUnitName})
	var duration = time.Since(now)

	switch {
	case err != nil:
		err = errors.Wrapf(err, "could not list '%s' systemd unit", agentSystemDUnitName)
		status = KO
	case len(units) == 0:
		err = fmt.Errorf("systemd unit '%s' not found", agentSystemDUnitName)
		status = KO
	case units[0].ActiveState != "active":
		err = fmt.Errorf("systemd unit '%s' is not active", agentSystemDUnitName)
		status = KO
	}

	return jaegerReport{
		Name:     name,
		Duration: duration.String(),
		Status:   status.String(),
		Error:    str(err),
	}
}

func (m *JaegerModule) jaegerCollectorPing() jaegerReport {
	var name = "ping collector"
	var status = OK

	// Query jaeger collector health check URL
	var now = time.Now()
	var res, err = m.httpClient.Get(fmt.Sprintf("http://%s", m.collectorHealthHostPort))
	var duration = time.Since(now)

	switch {
	case err != nil:
		err = errors.Wrap(err, "could not query jaeger collector health check service")
		status = KO
	case res.StatusCode != 204:
		err = errors.Wrapf(err, "jaeger health check service returned invalid status code: %v", res.StatusCode)
		status = KO
	default:
		status = OK
	}

	return jaegerReport{
		Name:     name,
		Duration: duration.String(),
		Status:   status.String(),
		Error:    str(err),
	}
}
