package interactors

import (
	"github.com/CESARBR/knot-babeltower/pkg/entities"
	"github.com/CESARBR/knot-babeltower/pkg/logging"
	"github.com/CESARBR/knot-babeltower/pkg/network"
)

// CreateUser to interact to user
type CreateUser struct {
	logger    logging.Logger
	userProxy network.UserProxy
}

// NewCreateUser contructs the interactor
func NewCreateUser(logger logging.Logger, userProxy network.UserProxy) *CreateUser {
	return &CreateUser{logger, userProxy}
}

// Execute runs the use case
func (cui *CreateUser) Execute(user entities.User) (err error) {
	cui.logger.Debug("Executing Create User interactor")

	err = cui.userProxy.SendCreateUser(user)
	if err != nil {
		cui.logger.Errorf("Send request error: %s", err.Error())
	}

	return err
}
