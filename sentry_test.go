package common_test

//go:generate mockgen --build_flags=--mod=mod -destination=./mock/sentry.go -package=mock -mock_names=SentryClient=SentryClient github.com/cloudtrust/common-healthcheck SentryClient

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/cloudtrust/common-healthcheck"
	mock "github.com/cloudtrust/common-healthcheck/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type sentryReport struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

func TestSentryDisabled(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSentry = mock.NewSentryClient(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer s.Close()

	var (
		enabled = false
		m       = NewSentryModule(mockSentry, s.Client(), enabled)
	)

	var jsonReport, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []sentryReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "sentry", r.Name)
	assert.Equal(t, "Deactivated", r.Status)
	assert.Zero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestSentryPing(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSentry = mock.NewSentryClient(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer s.Close()

	var (
		enabled = true
		url     = s.URL[7:]
		m       = NewSentryModule(mockSentry, s.Client(), enabled)
	)

	mockSentry.EXPECT().URL().Return(fmt.Sprintf("http://a:b@%s/api/1/store/", url)).Times(1)
	var jsonReport, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []sentryReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "ping", r.Name)
	assert.Equal(t, "OK", r.Status)
	assert.NotZero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestSentryAllChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSentry = mock.NewSentryClient(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer s.Close()

	var (
		enabled = true
		url     = s.URL[7:]
		m       = NewSentryModule(mockSentry, s.Client(), enabled)
	)

	mockSentry.EXPECT().URL().Return(fmt.Sprintf("http://a:b@%s/api/1/store/", url)).Times(1)
	var jsonReport, err = m.HealthCheck(context.Background(), "")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []sentryReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "ping", r.Name)
	assert.Equal(t, "OK", r.Status)
	assert.NotZero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestSentryFailure(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSentry = mock.NewSentryClient(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ko"))
	}))
	defer s.Close()

	var (
		enabled = true
		url     = s.URL[7:]
		m       = NewSentryModule(mockSentry, s.Client(), enabled)
	)

	mockSentry.EXPECT().URL().Return(fmt.Sprintf("http://a:b@%s/api/1/store/", url)).Times(1)
	var jsonReport, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []sentryReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "ping", r.Name)
	assert.Equal(t, "KO", r.Status)
	assert.NotZero(t, r.Duration)
	assert.NotZero(t, r.Error)
}

func TestSentryUnkownHealthCheck(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSentry = mock.NewSentryClient(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ko"))
	}))
	defer s.Close()

	var (
		enabled         = true
		m               = NewSentryModule(mockSentry, s.Client(), enabled)
		healthCheckName = "unknown"
	)

	var f = func() {
		m.HealthCheck(context.Background(), healthCheckName)
	}
	assert.Panics(t, f)
}
