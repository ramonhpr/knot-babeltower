package interactors

import (
	"testing"

	"github.com/CESARBR/knot-babeltower/pkg/entities"
)

type FakeCreateUserLogger struct {
}

type FakeUserProxy struct {
}

type FakeUserProxyWithSideEffect struct {
}

type ErrorMock struct {
	msg string
}

func (fl *FakeCreateUserLogger) Info(...interface{}) {}

func (fl *FakeCreateUserLogger) Infof(string, ...interface{}) {}

func (fl *FakeCreateUserLogger) Debug(...interface{}) {}

func (fl *FakeCreateUserLogger) Warn(...interface{}) {}

func (fl *FakeCreateUserLogger) Error(...interface{}) {}

func (fl *FakeCreateUserLogger) Errorf(string, ...interface{}) {}

func (fup *FakeUserProxy) SendCreateUser(user entities.User) (err error) {
	return nil
}

func (em *ErrorMock) Error() string {
	return em.msg
}

func (fup *FakeUserProxyWithSideEffect) SendCreateUser(user entities.User) error {
	return &ErrorMock{msg: "Error mocked"}
}

func shouldReturnNoError() error {
	fakeLogger := &FakeCreateUserLogger{}
	fakeUserProxy := &FakeUserProxy{}
	createUserInteractor := NewCreateUser(fakeLogger, fakeUserProxy)
	return createUserInteractor.Execute(entities.User{Email: "fake@email.com", Password: "123"})
}

func shouldRaiseError() error {
	fakeLogger := &FakeCreateUserLogger{}
	fakeUserProxy := &FakeUserProxyWithSideEffect{}
	createUserInteractor := NewCreateUser(fakeLogger, fakeUserProxy)
	user := entities.User{Email: "fake@email.com", Password: "123"}

	return createUserInteractor.Execute(user)
}

func TestCreateUser(t *testing.T) {
	err := shouldReturnNoError()
	if err != nil {
		t.Errorf("Create User failed. Error: %s", err)
	}

	err = shouldRaiseError()
	if err == nil {
		t.Errorf("Create User should raise error. Error: %s", err)
	}

	t.Logf("Create user ok")
}
