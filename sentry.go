package common

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/go-kit/kit/log"
)

// SentryModule is the health check module for sentry.
type SentryModule struct {
	sentry     SentryClient
	httpClient sentryHTTPClient
	enabled    bool
}

// sentryClient is the interface of the sentry client.
type SentryClient interface {
	URL() string
}



// sentryHTTPClient is the interface of the http client.
type sentryHTTPClient interface {
	Get(string) (*http.Response, error)
}

// NewSentryModule returns the sentry health module.
func NewSentryModule(sentry SentryClient, httpClient sentryHTTPClient, enabled bool) *SentryModule {
	return &SentryModule{
		sentry:     sentry,
		httpClient: httpClient,
		enabled:    enabled,
	}
}

// SentryReport is the health report returned by the sentry module.
type SentryReport struct {
	Name     string
	Duration time.Duration
	Status   Status
	Error    error
}


func (i *SentryReport) MarshalJSON() ([]byte, error) {
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

// HealthChecks executes all health checks for Sentry.
func (m *SentryModule) HealthChecks(context.Context) []SentryReport {
	var reports = []SentryReport{}
	reports = append(reports, m.sentryPingCheck())
	return reports
}

func (m *SentryModule) sentryPingCheck() SentryReport {
	var healthCheckName = "ping"

	if !m.enabled {
		return SentryReport{
			Name:   healthCheckName,
			Status: Deactivated,
		}
	}

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


// Logging middleware at module level.
type sentryModuleLoggingMW struct {
	logger log.Logger
	next   SentryHealthChecker
}

// SentryHealthChecker is the interface of the sentry health check module.
type SentryHealthChecker interface {
	HealthChecks(context.Context) []SentryReport
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

// sentryModuleLoggingMW implements Module.
func (m *sentryModuleLoggingMW) HealthChecks(ctx context.Context) []SentryReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}
