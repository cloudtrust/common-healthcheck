package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// FlakiCheckNames contains the list of all valid tests names.
var FlakiCheckNames = map[string]struct{}{
	"":       struct{}{},
	"nextid": struct{}{},
}

// NewFlakiModule returns the Flaki health module.
func NewFlakiModule(client FlakiClient, enabled bool) *FlakiModule {
	return &FlakiModule{
		flakiClient: client,
		enabled:     enabled,
	}
}

// FlakiModule is the health check module for Flaki.
type FlakiModule struct {
	flakiClient FlakiClient
	enabled     bool
}

// FlakiClient is the interface of the Flaki client.
type FlakiClient interface {
	NextID(context.Context) (string, error)
}

type flakiReport struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

// HealthCheck executes the desired influx health check.
func (m *FlakiModule) HealthCheck(_ context.Context, name string) (json.RawMessage, error) {
	if !m.enabled {
		return json.MarshalIndent(influxReport{Name: "flaki", Status: Deactivated.String()}, "", "  ")
	}

	switch name {
	case "":
		var reports []flakiReport
		reports = append(reports, m.nextID())
		return json.MarshalIndent(reports, "", "  ")
	case "ping":
		return json.MarshalIndent(m.nextID(), "", "  ")
	default:
		// Should not happen: there is a middleware validating the inputs name.
		panic(fmt.Sprintf("Unknown influx health check name: %v", name))
	}
}

func (m *FlakiModule) nextID() flakiReport {
	var name = "nextid"
	var status = OK

	var now = time.Now()
	var _, err = m.flakiClient.NextID(context.Background())
	var duration = time.Since(now)

	if err != nil {
		status = KO
		err = errors.Wrap(err, "could not get ID from flaki")
	}

	return flakiReport{
		Name:     name,
		Duration: duration.String(),
		Status:   status.String(),
		Error:    str(err),
	}
}
