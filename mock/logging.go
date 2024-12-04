// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/go-kit/kit/log (interfaces: Logger)
//
// Generated by this command:
//
//	mockgen --build_flags=--mod=mod -destination=./mock/logging.go -package=mock -mock_names=Logger=Logger github.com/go-kit/kit/log Logger
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// Logger is a mock of Logger interface.
type Logger struct {
	ctrl     *gomock.Controller
	recorder *LoggerMockRecorder
	isgomock struct{}
}

// LoggerMockRecorder is the mock recorder for Logger.
type LoggerMockRecorder struct {
	mock *Logger
}

// NewLogger creates a new mock instance.
func NewLogger(ctrl *gomock.Controller) *Logger {
	mock := &Logger{ctrl: ctrl}
	mock.recorder = &LoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Logger) EXPECT() *LoggerMockRecorder {
	return m.recorder
}

// Log mocks base method.
func (m *Logger) Log(keyvals ...any) error {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range keyvals {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Log", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Log indicates an expected call of Log.
func (mr *LoggerMockRecorder) Log(keyvals ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Log", reflect.TypeOf((*Logger)(nil).Log), keyvals...)
}