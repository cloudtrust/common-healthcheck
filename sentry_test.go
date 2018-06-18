package common_test

//go:generate mockgen -destination=./mock/sentry.go -package=mock -mock_names=SentryClient=SentryClient,SentryHealthChecker=SentryHealthChecker github.com/cloudtrust/common-healthcheck SentryClient,SentryHealthChecker
//go:generate mockgen -destination=./mock/logging.go -package=mock -mock_names=Logger=Logger github.com/go-kit/kit/log Logger

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	. "github.com/cloudtrust/common-healthcheck"
	mock "github.com/cloudtrust/common-healthcheck/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSentryHealthChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSentry = mock.NewSentryClient(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer s.Close()

	var m = NewSentryModule(mockSentry, s.Client(), true)

	mockSentry.EXPECT().URL().Return(fmt.Sprintf("http://a:b@%s/api/1/store/", s.URL[7:])).Times(1)
	var report = m.HealthChecks(context.Background())[0]
	assert.Equal(t, "ping", report.Name)
	assert.NotZero(t, report.Duration)
	assert.Equal(t, OK, report.Status)
	assert.Zero(t, report.Error)
}

func TestNoopSentryHealthChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSentry = mock.NewSentryClient(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer s.Close()

	var m = NewSentryModule(mockSentry, s.Client(), false)

	var report = m.HealthChecks(context.Background())[0]
	assert.Equal(t, "sentry", report.Name)
	assert.Zero(t, report.Duration)
	assert.Equal(t, Deactivated, report.Status)
	assert.Zero(t, report.Error)
}

func TestSentryModuleLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSentry = mock.NewSentryClient(mockCtrl)
	var mockLogger = mock.NewLogger(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer s.Close()

	var module = NewSentryModule(mockSentry, s.Client(), false)
	var m = MakeSentryModuleLoggingMW(mockLogger)(module)

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	mockLogger.EXPECT().Log("unit", "HealthChecks", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.HealthChecks(ctx)

	// Without correlation ID.
	var f = func() {
		m.HealthChecks(context.Background())
	}
	assert.Panics(t, f)
}

func TestSentryReportMarshalJSON(t *testing.T) {
	var report = &SentryReport{
		Name:     "Sentry",
		Duration: 1 * time.Second,
		Status:   OK,
		Error:    fmt.Errorf("Error"),
	}

	json, err := report.MarshalJSON()

	assert.Nil(t, err)
	assert.Equal(t, "{\"name\":\"Sentry\",\"duration\":\"1s\",\"status\":\"OK\",\"error\":\"Error\"}", string(json))
}
