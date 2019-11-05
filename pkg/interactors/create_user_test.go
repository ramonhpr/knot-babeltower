package interactors

import (
	"testing"

	"github.com/CESARBR/knot-babeltower/pkg/entities"
)

type FakeCreateUserLogger struct {
}

func (fl *FakeCreateUserLogger) Info(...interface{}) {}

func (fl *FakeCreateUserLogger) Infof(string, ...interface{}) {}

func (fl *FakeCreateUserLogger) Debug(...interface{}) {}

func (fl *FakeCreateUserLogger) Warn(...interface{}) {}

func (fl *FakeCreateUserLogger) Error(...interface{}) {}

func (fl *FakeCreateUserLogger) Errorf(string, ...interface{}) {}

func TestCreateUser(t *testing.T) {
	fakeLogger := &FakeCreateUserLogger{}
	createUserInteractor := NewCreateUser(fakeLogger)

	err := createUserInteractor.Execute(entities.User{Email: "fake@email.com", Password: "123"}) // should return no error
	if err != nil {
		t.Error("Create user fail")
	}

	t.Logf("Create user ok")
}
