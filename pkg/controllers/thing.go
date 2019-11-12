package controllers

import (
	"net/http"

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

// Create handles the server request and calls CreateThingInteractor
func (tc *ThingController) Create(w http.ResponseWriter, r *http.Request) {
	tc.logger.Debug("Create thing")
	// TODO: parse request
	err := tc.createThing.Execute()
	if err != nil {
		tc.logger.Error(err)
		return
	}
}
