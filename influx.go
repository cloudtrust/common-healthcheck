package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// InfluxCheckNames contains the list of all valid tests names.
var InfluxCheckNames = map[string]struct{}{
	"":     struct{}{},
	"ping": struct{}{},
}

// NewInfluxModule returns the influx health module.
func NewInfluxModule(influx InfluxClient, enabled bool) *InfluxModule {
	return &InfluxModule{
		influx:  influx,
		enabled: enabled,
	}
}

// InfluxModule is the health check module for influx.
type InfluxModule struct {
	influx  InfluxClient
	enabled bool
}

// InfluxClient is the interface of the influx client.
type InfluxClient interface {
	Ping(timeout time.Duration) (time.Duration, string, error)
}

type influxReport struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

// HealthCheck executes the desired influx health check.
func (m *InfluxModule) HealthCheck(_ context.Context, name string) (json.RawMessage, error) {
	if !m.enabled {
		return json.MarshalIndent(influxReport{Name: "influx", Status: Deactivated.String()}, "", "  ")
	}

	switch name {
	case "":
		var reports []influxReport
		reports = append(reports, m.influxPing())
		return json.MarshalIndent(reports, "", "  ")
	case "ping":
		return json.MarshalIndent(m.influxPing(), "", "  ")
	default:
		// Should not happen: there is a middleware validating the inputs name.
		panic(fmt.Sprintf("Unknown influx health check name: %v", name))
	}
}

func (m *InfluxModule) influxPing() influxReport {
	var name = "ping"
	var status = OK

	var now = time.Now()
	var _, _, err = m.influx.Ping(5 * time.Second)
	var duration = time.Since(now)

	if err != nil {
		status = KO
		err = errors.Wrap(err, "could not ping influx")
	}

	return influxReport{
		Name:     name,
		Duration: duration.String(),
		Status:   status.String(),
		Error:    str(err),
	}
}
