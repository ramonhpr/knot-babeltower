package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/CESARBR/knot-babeltower/pkg/logging"

	"github.com/CESARBR/knot-babeltower/pkg/entities"
	"github.com/CESARBR/knot-babeltower/pkg/interactors"
)

// UserController represents the controller for user
type UserController struct {
	logger               logging.Logger
	createUserInteractor *interactors.CreateUser
}

// NewUserController constructs the controller
func NewUserController(
	logger logging.Logger,
	createUserInteractor *interactors.CreateUser) *UserController {
	return &UserController{logger, createUserInteractor}
}

func (uc *UserController) writeResponse(w http.ResponseWriter, status int, err string) {
	js, jsonErr := json.Marshal(entities.StatusResponse{Code: status, Message: err})
	if jsonErr != nil {
		uc.logger.Errorf("Unable to marshal json: %s", jsonErr)
		return
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	_, writeErr := w.Write(js)
	if writeErr != nil {
		uc.logger.Errorf("Unable to write to connection HTTP: %s", writeErr)
		return
	}
}

// Create handles the server request and calls CreateUserInteractor
func (uc *UserController) Create(w http.ResponseWriter, r *http.Request) {
	var err error
	var errStr string
	var status int
	var user entities.User
	var decoder *json.Decoder

	uc.logger.Debug("Handle request to create user")

	decoder = json.NewDecoder(r.Body)

	err = decoder.Decode(&user)
	if err != nil {
		uc.logger.Errorf("Invalid request format: %s", err)
		status = http.StatusUnprocessableEntity
		errStr = err.Error()
		goto done
	}

	err = uc.createUserInteractor.Execute(user)
	if err != nil {
		uc.logger.Errorf("Response error: %s", err)
		status = http.StatusInternalServerError
		errStr = err.Error()
	} else {
		uc.logger.Infof("User %s created", user.Email)
		status = http.StatusCreated
	}
done:
	uc.writeResponse(w, status, errStr)
}
