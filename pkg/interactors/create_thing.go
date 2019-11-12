package interactors

import (
	"github.com/CESARBR/knot-babeltower/pkg/logging"
)

// CreateThing to interact to thing
type CreateThing struct {
	logger logging.Logger
}

// NewCreateThing contructs the interactor
func NewCreateThing(logger logging.Logger) *CreateThing {
	return &CreateThing{logger}
}

// Execute runs the use case
func (ct *CreateThing) Execute() error {
	ct.logger.Debug("Executing Create thing interactor")
	return nil
}
