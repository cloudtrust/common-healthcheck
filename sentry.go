package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// SentryCheckNames contains the list of all valid tests names.
var SentryCheckNames = map[string]struct{}{
	"":     struct{}{},
	"ping": struct{}{},
}

// NewSentryModule returns the sentry health module.
func NewSentryModule(sentry SentryClient, httpClient HTTPClient, enabled bool) *SentryModule {
	return &SentryModule{
		sentry:     sentry,
		httpClient: httpClient,
		enabled:    enabled,
	}
}

// SentryModule is the health check module for sentry.
type SentryModule struct {
	sentry     SentryClient
	httpClient HTTPClient
	enabled    bool
}

// SentryClient is the interface of the sentry client.
type SentryClient interface {
	URL() string
}

type sentryReport struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

// HealthCheck executes the desired influx health check.
func (m *SentryModule) HealthCheck(_ context.Context, name string) (json.RawMessage, error) {
	if !m.enabled {
		return json.MarshalIndent([]influxReport{{Name: "sentry", Status: Deactivated.String()}}, "", "  ")
	}

	var reports []sentryReport
	switch name {
	case "":
		reports = append(reports, m.sentryPing())
	case "ping":
		reports = append(reports, m.sentryPing())
	default:
		// Should not happen: there is a middleware validating the inputs name.
		panic(fmt.Sprintf("Unknown sentry health check name: %v", name))
	}

	return json.MarshalIndent(reports, "", "  ")
}

func (m *SentryModule) sentryPing() sentryReport {
	var name = "ping"
	var status = OK

	// Get Sentry health status.
	var now = time.Now()
	var err = m.getSentryHealth()
	var duration = time.Since(now)

	if err != nil {
		status = KO
		err = errors.Wrap(err, "could not ping sentry")
	}

	return sentryReport{
		Name:     name,
		Duration: duration.String(),
		Status:   status.String(),
		Error:    str(err),
	}
}

func (m *SentryModule) getSentryHealth() error {
	// Build sentry health url from sentry dsn. The health url is <sentryURL>/_health
	var dsn = m.sentry.URL()

	var url string
	if idx := strings.LastIndex(dsn, "/api/"); idx != -1 {
		url = fmt.Sprintf("%s/_health", dsn[:idx])
	}

	// Query sentry health endpoint.
	var res *http.Response
	{
		var err error
		res, err = m.httpClient.Get(url)
		if err != nil {
			return err
		}
		if res != nil {
			defer res.Body.Close()
		}
	}

	// Chesk response status.
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("http response status code: %v", res.Status)
	}

	// Chesk response body. The sentry health endpoint returns "ok" when there is no issue.
	var response []byte
	{
		var err error
		response, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
	}

	if strings.Compare(string(response), "ok") == 0 {
		return nil
	}

	return fmt.Errorf("response should be 'ok' but is: %v", string(response))
}
