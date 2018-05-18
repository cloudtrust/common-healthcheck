package common_test

//go:generate mockgen -destination=./mock/flaki.go -package=mock -mock_names=FlakiClient=FlakiClient  github.com/cloudtrust/common-healthcheck FlakiClient
//go:generate mockgen -destination=./mock/logging.go -package=mock -mock_names=Logger=Logger github.com/go-kit/kit/log Logger

import (
	"context"
	"testing"
	"strconv"
	"math/rand"
	"time"
	 "fmt"

	. "github.com/cloudtrust/common-healthcheck"
	mock "github.com/cloudtrust/common-healthcheck/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFlakiHealthChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish() 
	var mockFlakiClient = mock.NewFlakiClient(mockCtrl)

	var m = NewFlakiModule(mockFlakiClient)

	mockFlakiClient.EXPECT().NextValidID().Return("00000-0000-0000", nil).Times(1)
	var report = m.HealthChecks(context.Background())[0]
	assert.Equal(t, "Flaki ID generation", report.Name)
	assert.NotZero(t, report.Duration)
	assert.Equal(t, OK, report.Status)
	assert.Zero(t, report.Error)
}

func TestFlakiFailureHealthChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish() 
	var mockFlakiClient = mock.NewFlakiClient(mockCtrl)

	var m = NewFlakiModule(mockFlakiClient)

	mockFlakiClient.EXPECT().NextValidID().Return("", fmt.Errorf("Error")).Times(1)
	var report = m.HealthChecks(context.Background())[0]
	assert.Equal(t, "Flaki ID generation", report.Name)
	assert.NotZero(t, report.Duration)
	assert.Equal(t, KO, report.Status)
	assert.Equal(t, "could not query flaki service: Error", report.Error.Error())
}

func TestFlakiModuleLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockFlakiClient = mock.NewFlakiClient(mockCtrl)
	var mockLogger = mock.NewLogger(mockCtrl)

	var module = NewFlakiModule(mockFlakiClient)
	var m = MakeFlakiModuleLoggingMW(mockLogger)(module)

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	
	mockFlakiClient.EXPECT().NextValidID().Return("000-000-000", nil).Times(1)
	mockLogger.EXPECT().Log("unit", "HealthChecks", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.HealthChecks(ctx)
 
	mockFlakiClient.EXPECT().NextValidID().Return("000-000-000", nil).Times(1)
	// Without correlation ID.
	var f = func() {
		m.HealthChecks(context.Background())
	}
	assert.Panics(t, f)
}

func TestFlakiReportMarshalJSON(t *testing.T) {
	var report = &JaegerReport{
		Name:     "Flaki",
		Duration: 1 * time.Second,
		Status:   OK,
		Error:    fmt.Errorf("Error"),
	}

	json, err := report.MarshalJSON()

	assert.Nil(t, err)
	assert.Equal(t, "{\"name\":\"Flaki\",\"duration\":\"1s\",\"status\":\"OK\",\"error\":\"Error\"}", string(json))
}