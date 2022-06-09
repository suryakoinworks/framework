// Code generated by mockery v2.12.3. DO NOT EDIT.

package mocks

import (
	configs "github.com/KejawenLab/bima/v2/configs"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// Model is an autogenerated mock type for the Model type
type Model struct {
	mock.Mock
}

// IsSoftDelete provides a mock function with given fields:
func (_m *Model) IsSoftDelete() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// SetCreatedAt provides a mock function with given fields: _a0
func (_m *Model) SetCreatedAt(_a0 time.Time) {
	_m.Called(_a0)
}

// SetCreatedBy provides a mock function with given fields: user
func (_m *Model) SetCreatedBy(user *configs.User) {
	_m.Called(user)
}

// SetDeletedAt provides a mock function with given fields: _a0
func (_m *Model) SetDeletedAt(_a0 time.Time) {
	_m.Called(_a0)
}

// SetDeletedBy provides a mock function with given fields: user
func (_m *Model) SetDeletedBy(user *configs.User) {
	_m.Called(user)
}

// SetSyncedAt provides a mock function with given fields: _a0
func (_m *Model) SetSyncedAt(_a0 time.Time) {
	_m.Called(_a0)
}

// SetUpdatedAt provides a mock function with given fields: _a0
func (_m *Model) SetUpdatedAt(_a0 time.Time) {
	_m.Called(_a0)
}

// SetUpdatedBy provides a mock function with given fields: user
func (_m *Model) SetUpdatedBy(user *configs.User) {
	_m.Called(user)
}

// TableName provides a mock function with given fields:
func (_m *Model) TableName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

type NewModelT interface {
	mock.TestingT
	Cleanup(func())
}

// NewModel creates a new instance of Model. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewModel(t NewModelT) *Model {
	mock := &Model{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
