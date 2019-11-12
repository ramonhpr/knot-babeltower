package interactors

import (
	"testing"

	"github.com/CESARBR/knot-babeltower/pkg/entities"
)

type FakeCreateThingLogger struct {
}

func (fl *FakeCreateThingLogger) Info(...interface{}) {}

func (fl *FakeCreateThingLogger) Infof(string, ...interface{}) {}

func (fl *FakeCreateThingLogger) Debug(...interface{}) {}

func (fl *FakeCreateThingLogger) Warn(...interface{}) {}

func (fl *FakeCreateThingLogger) Error(...interface{}) {}

func (fl *FakeCreateThingLogger) Errorf(string, ...interface{}) {}

func TestCreateThing(t *testing.T) {
	testCases := []struct {
		name       string
		fakeLogger *FakeCreateThingLogger
		expected   error
	}{
		{
			"shouldReturnNoError",
			&FakeCreateThingLogger{},
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createThingInteractor := NewCreateThing(tc.fakeLogger)
			err := createThingInteractor.Execute(entities.Thing{Id: "fakeId", Name: "FakeName", Schema: map[string]interface{}{}})
			if err != tc.expected {
				t.Errorf("Create Thing failed. Error: %s", err)
				return
			}

			t.Logf("Create thing ok")
		})
	}

}
