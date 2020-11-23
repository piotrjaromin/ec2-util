// Code generated by MockGen. DO NOT EDIT.
// Source: mongo/init_replica.go

// Package mongomocks is a generated GoMock package.
package mongomocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockmongoSession is a mock of mongoSession interface
type MockmongoSession struct {
	ctrl     *gomock.Controller
	recorder *MockmongoSessionMockRecorder
}

// MockmongoSessionMockRecorder is the mock recorder for MockmongoSession
type MockmongoSessionMockRecorder struct {
	mock *MockmongoSession
}

// NewMockmongoSession creates a new mock instance
func NewMockmongoSession(ctrl *gomock.Controller) *MockmongoSession {
	mock := &MockmongoSession{ctrl: ctrl}
	mock.recorder = &MockmongoSessionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockmongoSession) EXPECT() *MockmongoSessionMockRecorder {
	return m.recorder
}

// Run mocks base method
func (m *MockmongoSession) Run(cmd, result interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", cmd, result)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run
func (mr *MockmongoSessionMockRecorder) Run(cmd, result interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockmongoSession)(nil).Run), cmd, result)
}
