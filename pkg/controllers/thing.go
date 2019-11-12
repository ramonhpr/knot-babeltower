package controllers

import (
	"net/http"

	"github.com/CESARBR/knot-babeltower/pkg/logging"
)

// ThingController represents the controller for thing
type ThingController struct {
	logger logging.Logger
}

// NewThingController constructs the controller
func NewThingController(logger logging.Logger) *ThingController {
	return &ThingController{logger}
}

// Create handles the server request and calls CreateThingInteractor
func (tc *ThingController) Create(w http.ResponseWriter, r *http.Request) {
	tc.logger.Debug("Create thing")
	// TODO: parse request
}
