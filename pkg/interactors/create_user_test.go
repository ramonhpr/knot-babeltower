package interactors

import "testing"

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

	createUserInteractor.Execute() // should return no error

	t.Logf("Create user ok")
}
