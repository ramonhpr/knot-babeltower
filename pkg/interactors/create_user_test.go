package interactors

import (
	"testing"

	"github.com/CESARBR/knot-babeltower/pkg/entities"
)

type FakeCreateUserLogger struct {
}

type FakeUserProxy struct {
	returnValue error
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
	return fup.returnValue
}

func (em *ErrorMock) Error() string {
	return em.msg
}

func TestCreateUser(t *testing.T) {
	testCases := []struct {
		name          string
		fakeLogger    *FakeCreateUserLogger
		fakeUserProxy *FakeUserProxy
	}{
		{
			"shouldReturnNoError",
			&FakeCreateUserLogger{},
			&FakeUserProxy{returnValue: nil},
		},
		{
			"shouldRaiseError",
			&FakeCreateUserLogger{},
			&FakeUserProxy{returnValue: &ErrorMock{msg: "Error mocked"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createUserInteractor := NewCreateUser(tc.fakeLogger, tc.fakeUserProxy)
			err := createUserInteractor.Execute(entities.User{Email: "fake@email.com", Password: "123"})
			if err != tc.fakeUserProxy.returnValue {
				t.Errorf("Create User failed. Error: %s", err)
				return
			}
			t.Logf("Create user ok")
		})
	}

}
