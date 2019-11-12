package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/CESARBR/knot-babeltower/pkg/entities"

	"github.com/CESARBR/knot-babeltower/pkg/interactors"

	"github.com/CESARBR/knot-babeltower/pkg/logging"
)

// ThingController represents the controller for thing
type ThingController struct {
	logger      logging.Logger
	createThing *interactors.CreateThing
}

// NewThingController constructs the controller
func NewThingController(logger logging.Logger, createThing *interactors.CreateThing) *ThingController {
	return &ThingController{logger, createThing}
}

func (tc *ThingController) writeResponse(w http.ResponseWriter, status int, err string) {
	js, jsonErr := json.Marshal(entities.StatusResponse{Code: status, Message: err})
	if jsonErr != nil {
		tc.logger.Errorf("Unable to marshal json: %s", jsonErr)
		return
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	_, writeErr := w.Write(js)
	if writeErr != nil {
		tc.logger.Errorf("Unable to write to connection HTTP: %s", writeErr)
		return
	}
}

// Create handles the server request and calls CreateThingInteractor
func (tc *ThingController) Create(w http.ResponseWriter, r *http.Request) {
	var err error
	var errStr string
	var status int
	var thing entities.Thing
	var decoder *json.Decoder

	tc.logger.Debug("Create thing")

	decoder = json.NewDecoder(r.Body)

	err = decoder.Decode(&thing)
	if err != nil {
		tc.logger.Errorf("Invalid request format: %s", err)
		status = http.StatusUnprocessableEntity
		errStr = err.Error()
		goto done
	}

	err = tc.createThing.Execute(thing)
	if err != nil {
		status = http.StatusUnprocessableEntity
		errStr = err.Error()
		goto done
	} else {
		tc.logger.Infof("User %s created", thing.Id)
		status = http.StatusCreated
	}

done:
	tc.writeResponse(w, status, errStr)
}
