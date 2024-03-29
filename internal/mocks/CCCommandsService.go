// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	commands "chemin-du-local.bzh/graphql/internal/services/commands"
	mock "github.com/stretchr/testify/mock"

	model "chemin-du-local.bzh/graphql/graph/model"

	options "go.mongodb.org/mongo-driver/mongo/options"

	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

// CCCommandsService is an autogenerated mock type for the CCCommandsService type
type CCCommandsService struct {
	mock.Mock
}

// Create provides a mock function with given fields: commerceCommandID, input
func (_m *CCCommandsService) Create(commerceCommandID primitive.ObjectID, input model.NewCCCommand) (*commands.CCCommand, error) {
	ret := _m.Called(commerceCommandID, input)

	var r0 *commands.CCCommand
	if rf, ok := ret.Get(0).(func(primitive.ObjectID, model.NewCCCommand) *commands.CCCommand); ok {
		r0 = rf(commerceCommandID, input)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commands.CCCommand)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(primitive.ObjectID, model.NewCCCommand) error); ok {
		r1 = rf(commerceCommandID, input)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetById provides a mock function with given fields: id
func (_m *CCCommandsService) GetById(id string) (*commands.CCCommand, error) {
	ret := _m.Called(id)

	var r0 *commands.CCCommand
	if rf, ok := ret.Get(0).(func(string) *commands.CCCommand); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commands.CCCommand)
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
func (_m *CCCommandsService) GetFiltered(filter interface{}, opts *options.FindOptions) ([]commands.CCCommand, error) {
	ret := _m.Called(filter, opts)

	var r0 []commands.CCCommand
	if rf, ok := ret.Get(0).(func(interface{}, *options.FindOptions) []commands.CCCommand); ok {
		r0 = rf(filter, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]commands.CCCommand)
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

// GetForCommmerceCommand provides a mock function with given fields: commerceCommandID
func (_m *CCCommandsService) GetForCommmerceCommand(commerceCommandID string) ([]commands.CCCommand, error) {
	ret := _m.Called(commerceCommandID)

	var r0 []commands.CCCommand
	if rf, ok := ret.Get(0).(func(string) []commands.CCCommand); ok {
		r0 = rf(commerceCommandID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]commands.CCCommand)
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

// GetProducts provides a mock function with given fields: cccommandID
func (_m *CCCommandsService) GetProducts(cccommandID string) ([]*model.CCProduct, error) {
	ret := _m.Called(cccommandID)

	var r0 []*model.CCProduct
	if rf, ok := ret.Get(0).(func(string) []*model.CCProduct); ok {
		r0 = rf(cccommandID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.CCProduct)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(cccommandID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewCCCommandsService interface {
	mock.TestingT
	Cleanup(func())
}

// NewCCCommandsService creates a new instance of CCCommandsService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewCCCommandsService(t mockConstructorTestingTNewCCCommandsService) *CCCommandsService {
	mock := &CCCommandsService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
