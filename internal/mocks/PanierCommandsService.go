// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	commands "chemin-du-local.bzh/graphql/internal/services/commands"
	mock "github.com/stretchr/testify/mock"

	model "chemin-du-local.bzh/graphql/graph/model"

	options "go.mongodb.org/mongo-driver/mongo/options"

	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

// PanierCommandsService is an autogenerated mock type for the PanierCommandsService type
type PanierCommandsService struct {
	mock.Mock
}

// Create provides a mock function with given fields: commerceCommandID, input
func (_m *PanierCommandsService) Create(commerceCommandID primitive.ObjectID, input model.NewPanierCommand) (*commands.PanierCommand, error) {
	ret := _m.Called(commerceCommandID, input)

	var r0 *commands.PanierCommand
	if rf, ok := ret.Get(0).(func(primitive.ObjectID, model.NewPanierCommand) *commands.PanierCommand); ok {
		r0 = rf(commerceCommandID, input)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commands.PanierCommand)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(primitive.ObjectID, model.NewPanierCommand) error); ok {
		r1 = rf(commerceCommandID, input)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetById provides a mock function with given fields: id
func (_m *PanierCommandsService) GetById(id string) (*commands.PanierCommand, error) {
	ret := _m.Called(id)

	var r0 *commands.PanierCommand
	if rf, ok := ret.Get(0).(func(string) *commands.PanierCommand); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commands.PanierCommand)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetFiltered provides a mock function with given fields: filter, opts
func (_m *PanierCommandsService) GetFiltered(filter interface{}, opts *options.FindOptions) ([]commands.PanierCommand, error) {
	ret := _m.Called(filter, opts)

	var r0 []commands.PanierCommand
	if rf, ok := ret.Get(0).(func(interface{}, *options.FindOptions) []commands.PanierCommand); ok {
		r0 = rf(filter, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]commands.PanierCommand)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(interface{}, *options.FindOptions) error); ok {
		r1 = rf(filter, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetForCommerceCommand provides a mock function with given fields: commerceCommandID
func (_m *PanierCommandsService) GetForCommerceCommand(commerceCommandID string) ([]commands.PanierCommand, error) {
	ret := _m.Called(commerceCommandID)

	var r0 []commands.PanierCommand
	if rf, ok := ret.Get(0).(func(string) []commands.PanierCommand); ok {
		r0 = rf(commerceCommandID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]commands.PanierCommand)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(commerceCommandID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewPanierCommandsService interface {
	mock.TestingT
	Cleanup(func())
}

// NewPanierCommandsService creates a new instance of PanierCommandsService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewPanierCommandsService(t mockConstructorTestingTNewPanierCommandsService) *PanierCommandsService {
	mock := &PanierCommandsService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}