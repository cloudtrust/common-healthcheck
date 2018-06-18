package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

// NewSentryModule returns the sentry health module.
func NewSentryModule(sentry SentryClient, httpClient sentryHTTPClient, enabled bool) *SentryModule {
	return &SentryModule{
		sentry:     sentry,
		httpClient: httpClient,
		enabled:    enabled,
	}
}

// SentryModule is the health check module for sentry.
type SentryModule struct {
	sentry     SentryClient
	httpClient sentryHTTPClient
	enabled    bool
}

// SentryClient is the interface of the sentry client.
type SentryClient interface {
	URL() string
}

// sentryHTTPClient is the interface of the http client.
type sentryHTTPClient interface {
	Get(string) (*http.Response, error)
}

// SentryReport is the health report returned by the sentry module.
type SentryReport struct {
	Name     string
	Duration time.Duration
	Status   Status
	Error    error
}

// MarshalJSON marshal the sentry report.
func (r *SentryReport) MarshalJSON() ([]byte, error) {
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

// HealthChecks executes all health checks for Sentry.
func (m *SentryModule) HealthChecks(context.Context) []SentryReport {
	if !m.enabled {
		return []SentryReport{{Name: "sentry", Status: Deactivated}}
	}

	var reports = []SentryReport{}
	reports = append(reports, m.sentryPingCheck())
	return reports
}

func (m *SentryModule) sentryPingCheck() SentryReport {
	var healthCheckName = "ping"

	var dsn = m.sentry.URL()

	// Get Sentry health status.
	var now = time.Now()
	var err = pingSentry(dsn, m.httpClient)
	var duration = time.Since(now)

	var hcErr error
	var s Status
	switch {
	case err != nil:
		hcErr = errors.Wrap(err, "could not ping sentry")
		s = KO
	default:
		s = OK
	}

	return SentryReport{
		Name:     healthCheckName,
		Duration: duration,
		Status:   s,
		Error:    hcErr,
	}
}

func pingSentry(dsn string, httpClient sentryHTTPClient) error {
	// Build sentry health url from sentry dsn. The health url is <sentryURL>/_health
	var url string
	if idx := strings.LastIndex(dsn, "/api/"); idx != -1 {
		url = fmt.Sprintf("%s/_health", dsn[:idx])
	}

	// Query sentry health endpoint.
	var res *http.Response
	{
		var err error
		res, err = httpClient.Get(url)
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

// MakeSentryModuleLoggingMW makes a logging middleware at module level.
func MakeSentryModuleLoggingMW(logger log.Logger) func(SentryHealthChecker) SentryHealthChecker {
	return func(next SentryHealthChecker) SentryHealthChecker {
		return &sentryModuleLoggingMW{
			logger: logger,
			next:   next,
		}
	}
}

// SentryHealthChecker is the interface of the sentry health check module.
type SentryHealthChecker interface {
	HealthChecks(context.Context) []SentryReport
}

// Logging middleware at module level.
type sentryModuleLoggingMW struct {
	logger log.Logger
	next   SentryHealthChecker
}

// sentryModuleLoggingMW implements SentryHealthChecker. There must be a key "correlation_id" with a string value in the context.
func (m *sentryModuleLoggingMW) HealthChecks(ctx context.Context) []SentryReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}
