// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import (
	http "net/http"

	grpc "google.golang.org/grpc"

	middlewares "github.com/KejawenLab/bima/v4/middlewares"

	mock "github.com/stretchr/testify/mock"
)

// Route is an autogenerated mock type for the Route type
type Route struct {
	mock.Mock
}

// Handle provides a mock function with given fields: w, r, params
func (_m *Route) Handle(w http.ResponseWriter, r *http.Request, params map[string]string) {
	_m.Called(w, r, params)
}

// Method provides a mock function with given fields:
func (_m *Route) Method() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Middlewares provides a mock function with given fields:
func (_m *Route) Middlewares() []middlewares.Middleware {
	ret := _m.Called()

	var r0 []middlewares.Middleware
	if rf, ok := ret.Get(0).(func() []middlewares.Middleware); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]middlewares.Middleware)
		}
	}

	return r0
}

// Path provides a mock function with given fields:
func (_m *Route) Path() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// SetClient provides a mock function with given fields: client
func (_m *Route) SetClient(client *grpc.ClientConn) {
	_m.Called(client)
}

type mockConstructorTestingTNewRoute interface {
	mock.TestingT
	Cleanup(func())
}

// NewRoute creates a new instance of Route. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRoute(t mockConstructorTestingTNewRoute) *Route {
	mock := &Route{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
