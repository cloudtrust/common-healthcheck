// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cloudtrust/common-healthcheck (interfaces: FlakiClient)
//
// Generated by this command:
//
//	mockgen --build_flags=--mod=mod -destination=./mock/flaki.go -package=mock -mock_names=FlakiClient=FlakiClient github.com/cloudtrust/common-healthcheck FlakiClient
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// FlakiClient is a mock of FlakiClient interface.
type FlakiClient struct {
	ctrl     *gomock.Controller
	recorder *FlakiClientMockRecorder
	isgomock struct{}
}

// FlakiClientMockRecorder is the mock recorder for FlakiClient.
type FlakiClientMockRecorder struct {
	mock *FlakiClient
}

// NewFlakiClient creates a new mock instance.
func NewFlakiClient(ctrl *gomock.Controller) *FlakiClient {
	mock := &FlakiClient{ctrl: ctrl}
	mock.recorder = &FlakiClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *FlakiClient) EXPECT() *FlakiClientMockRecorder {
	return m.recorder
}

// NextID mocks base method.
func (m *FlakiClient) NextID(arg0 context.Context) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NextID", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NextID indicates an expected call of NextID.
func (mr *FlakiClientMockRecorder) NextID(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NextID", reflect.TypeOf((*FlakiClient)(nil).NextID), arg0)
}
