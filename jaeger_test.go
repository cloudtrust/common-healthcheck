package common_test

//go:generate mockgen -destination=./mock/jaeger.go -package=mock -mock_names=SystemDConn=SystemDConn  github.com/cloudtrust/common-healthcheck SystemDConn

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
	"github.com/coreos/go-systemd/dbus"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestJaegerHealthChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSystemDConn = mock.NewSystemDConn(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer s.Close()

	var m = NewJaegerModule(mockSystemDConn, s.Client(), "jaeger-collector:14269", true)

	var units = []dbus.UnitStatus{{Name: "agent.service", ActiveState: "active"}}

	// HealthChecks
	{
		mockSystemDConn.EXPECT().ListUnitsByNames([]string{"agent.service"}).Return(units, nil).Times(1)
		var report = m.HealthChecks(context.Background())[0]
		assert.Equal(t, "jaeger agent systemd unit check", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, OK, report.Status)
		assert.Zero(t, report.Error)
	}

	// SystemD fail.
	{
		mockSystemDConn.EXPECT().ListUnitsByNames([]string{"agent.service"}).Return(nil, fmt.Errorf("fail")).Times(1)
		var report = m.HealthChecks(context.Background())[0]
		assert.Equal(t, "jaeger agent systemd unit check", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, KO, report.Status)
		assert.NotZero(t, report.Error)
	}
}

func TestNoopJaegerHealthChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSystemDConn = mock.NewSystemDConn(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer s.Close()

	var m = NewJaegerModule(mockSystemDConn, s.Client(), "jaeger-collector:14269", false)

	var report = m.HealthChecks(context.Background())[0]
	assert.Equal(t, "jaeger agent systemd unit check", report.Name)
	assert.Zero(t, report.Duration)
	assert.Equal(t, Deactivated, report.Status)
	assert.Zero(t, report.Error)
}

func TestJaegerModuleLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSystemDConn = mock.NewSystemDConn(mockCtrl)
	var mockLogger = mock.NewLogger(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer s.Close()

	var module = NewJaegerModule(mockSystemDConn, s.Client(), "jaeger-collector:14269", false)

	var m = MakeJaegerModuleLoggingMW(mockLogger)(module)

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

func TestJaegerReportMarshalJSON(t *testing.T) {
	var report = &JaegerReport{
		Name:     "Jaeger",
		Duration: 1 * time.Second,
		Status:   OK,
		Error:    fmt.Errorf("Error"),
	}

	json, err := report.MarshalJSON()

	assert.Nil(t, err)
	assert.Equal(t, "{\"name\":\"Jaeger\",\"duration\":\"1s\",\"status\":\"OK\",\"error\":\"Error\"}", string(json))
}
