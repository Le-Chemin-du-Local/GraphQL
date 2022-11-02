// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	model "chemin-du-local.bzh/graphql/graph/model"
	mock "github.com/stretchr/testify/mock"
)

// CommandsService is an autogenerated mock type for the CommandsService type
type CommandsService struct {
	mock.Mock
}

// GetUser provides a mock function with given fields: commandID
func (_m *CommandsService) GetUser(commandID string) (*model.User, error) {
	ret := _m.Called(commandID)

	var r0 *model.User
	if rf, ok := ret.Get(0).(func(string) *model.User); ok {
		r0 = rf(commandID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(commandID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewCommandsService interface {
	mock.TestingT
	Cleanup(func())
}

// NewCommandsService creates a new instance of CommandsService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewCommandsService(t mockConstructorTestingTNewCommandsService) *CommandsService {
	mock := &CommandsService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
