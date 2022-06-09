// Code generated by mockery v2.12.3. DO NOT EDIT.

package mocks

import (
	configs "github.com/KejawenLab/bima/v2/configs"
	mock "github.com/stretchr/testify/mock"
)

// Application is an autogenerated mock type for the Application type
type Application struct {
	mock.Mock
}

// IsBackground provides a mock function with given fields:
func (_m *Application) IsBackground() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Priority provides a mock function with given fields:
func (_m *Application) Priority() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// Run provides a mock function with given fields: servers
func (_m *Application) Run(servers []configs.Server) {
	_m.Called(servers)
}

type NewApplicationT interface {
	mock.TestingT
	Cleanup(func())
}

// NewApplication creates a new instance of Application. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewApplication(t NewApplicationT) *Application {
	mock := &Application{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
