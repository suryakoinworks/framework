// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"

	runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// Router is an autogenerated mock type for the Router type
type Router struct {
	mock.Mock
}

// Handle provides a mock function with given fields: _a0, server, client
func (_m *Router) Handle(_a0 context.Context, server *runtime.ServeMux, client *grpc.ClientConn) {
	_m.Called(_a0, server, client)
}

// Priority provides a mock function with given fields:
func (_m *Router) Priority() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

type mockConstructorTestingTNewRouter interface {
	mock.TestingT
	Cleanup(func())
}

// NewRouter creates a new instance of Router. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRouter(t mockConstructorTestingTNewRouter) *Router {
	mock := &Router{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
