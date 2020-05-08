package mocks

import (
	"errors"

	"github.com/stretchr/testify/mock"
)

// FakeController represents a mocking thing controller
type FakeController struct {
	mock.Mock
}

var errEmptyBody = errors.New("empty body")

// Register provides a mock function to not return error
func (f *FakeController) Register(body []byte, authorizationHeader string) error {
	if len(body) == 0 {
		return errEmptyBody
	}
	f.Called()
	return nil
}

// Unregister provides a mock function to not return error
func (f *FakeController) Unregister(body []byte, authorizationHeader string) error {
	if len(body) == 0 {
		return errEmptyBody
	}
	f.Called()
	return nil
}

// UpdateSchema provides a mock function to not return error
func (f *FakeController) UpdateSchema(body []byte, authorizationHeader string) error {
	if len(body) == 0 {
		return errEmptyBody
	}
	f.Called()
	return nil
}

// AuthDevice provides a mock function to not return error
func (f *FakeController) AuthDevice(body []byte, authorization string, replyTo, corrID string) error {
	f.Called()
	return nil
}

// ListDevices provides a mock function to not return error
func (f *FakeController) ListDevices(authorization string, replyTo, corrID string) error {
	f.Called()
	return nil
}

// PublishData provides a mock function to not return error
func (f *FakeController) PublishData(body []byte, authorization string) error {
	if len(body) == 0 {
		return errEmptyBody
	}
	f.Called()
	return nil
}

// RequestData provides a mock function to not return error
func (f *FakeController) RequestData(body []byte, authorization string) error {
	if len(body) == 0 {
		return errEmptyBody
	}
	f.Called()
	return nil
}

// UpdateData provides a mock function to not return error
func (f *FakeController) UpdateData(body []byte, authorization string) error {
	if len(body) == 0 {
		return errEmptyBody
	}
	f.Called()
	return nil
}
