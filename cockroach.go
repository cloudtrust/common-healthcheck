package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// CockroachCheckNames contains the list of all valid tests names.
var CockroachCheckNames = map[string]struct{}{
	"":     struct{}{},
	"ping": struct{}{},
}

// NewCockroachModule returns the cockroach health module.
func NewCockroachModule(cockroach CockroachClient, enabled bool) *CockroachModule {
	return &CockroachModule{
		cockroach: cockroach,
		enabled:   enabled,
	}
}

// CockroachModule is the health check module for cockroach.
type CockroachModule struct {
	cockroach CockroachClient
	enabled   bool
}

// CockroachClient is the interface of the cockroach client.
type CockroachClient interface {
	Ping() error
}

type cockroachReport struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

// HealthCheck executes the desired cockroach health check.
func (m *CockroachModule) HealthCheck(_ context.Context, name string) (json.RawMessage, error) {
	if !m.enabled {
		return json.MarshalIndent([]cockroachReport{{Name: "cockroach", Status: Deactivated.String()}}, "", "  ")
	}

	var reports []cockroachReport
	switch name {
	case "":
		reports = append(reports, m.cockroachPing())
	case "ping":
		reports = append(reports, m.cockroachPing())
	default:
		// Should not happen: there is a middleware validating the inputs name.
		panic(fmt.Sprintf("Unknown cockroach health check name: %v", name))
	}

	return json.MarshalIndent(reports, "", "  ")
}

func (m *CockroachModule) cockroachPing() cockroachReport {
	var name = "ping"
	var status = OK

	var now = time.Now()
	var err = m.cockroach.Ping()
	var duration = time.Since(now)

	if err != nil {
		status = KO
		err = errors.Wrap(err, "could not ping cockroach")
	}

	return cockroachReport{
		Name:     name,
		Duration: duration.String(),
		Status:   status.String(),
		Error:    str(err),
	}
}
