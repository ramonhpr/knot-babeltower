package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CESARBR/knot-babeltower/pkg/logging"

	"github.com/gorilla/mux"
)

// Controller defines which functions the controller should have
type Controller interface {
	Create(w http.ResponseWriter, r *http.Request)
}

// Health represents the service's health status
type Health struct {
	Status string `json:"status"`
}

type handler func(http.ResponseWriter, *http.Request)

// Server represents the HTTP server
type Server struct {
	port      int
	logger    logging.Logger
	ctlMapper map[string]map[string]handler
}

// NewServer creates a new server instance
func NewServer(port int, logger logging.Logger, userController, thingController Controller) Server {
	return Server{port, logger, map[string]map[string]handler{
		"/users": map[string]handler{
			"POST": userController.Create},
		"/things": map[string]handler{
			"POST": thingController.Create}}}
}

// Start starts the http server
func (s *Server) Start() {
	routers := s.createRouters()
	s.logger.Infof("Listening on %d", s.port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.logRequest(routers))
	if err != nil {
		s.logger.Error(err)
	}
}

func (s *Server) createRouters() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/healthcheck", s.healthcheckHandler)
	for route, methodMapper := range s.ctlMapper {
		for method, handlerCtl := range methodMapper {
			r.HandleFunc(route, handlerCtl).Methods(method)
		}
	}

	return r
}

func (s *Server) logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Infof("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func (s *Server) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response, _ := json.Marshal(&Health{Status: "online"})
	_, err := w.Write(response)
	if err != nil {
		s.logger.Errorf("Error sending response, %s\n", err)
	}
}
