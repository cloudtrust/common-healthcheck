// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cloudtrust/common-healthcheck (interfaces: CockroachClient)

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// CockroachClient is a mock of CockroachClient interface
type CockroachClient struct {
	ctrl     *gomock.Controller
	recorder *CockroachClientMockRecorder
}

// CockroachClientMockRecorder is the mock recorder for CockroachClient
type CockroachClientMockRecorder struct {
	mock *CockroachClient
}

// NewCockroachClient creates a new mock instance
func NewCockroachClient(ctrl *gomock.Controller) *CockroachClient {
	mock := &CockroachClient{ctrl: ctrl}
	mock.recorder = &CockroachClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *CockroachClient) EXPECT() *CockroachClientMockRecorder {
	return m.recorder
}

// Ping mocks base method
func (m *CockroachClient) Ping() error {
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping
func (mr *CockroachClientMockRecorder) Ping() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*CockroachClient)(nil).Ping))
}
