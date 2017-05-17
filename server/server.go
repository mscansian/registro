// Package server provides REST server implementation for the Service Registry.
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// NewServer returns a new server instance with the selected ListenAddr.
func NewServer(addr string) *Server {
	return &Server{
		ListenAddr:   addr,
		Applications: make([]*Application, 0),
	}
}

// Server represents a Service Register REST server.
// It can be used by calling Serve.
type Server struct {
	// ListenAddr is the address for the listening socket
	ListenAddr string

	// Applications holds the list of apps registered.
	Applications []*Application
}

// Serve start listening on ListenAddr for REST requests.
func (s *Server) Serve() error {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/registro/1.0/apps", s.listAppsHandler)
	router.HandleFunc("/registro/1.0/apps/{appName}", s.viewAppHandler)
	router.HandleFunc("/registro/1.0/apps/{appName}/{instanceId}", s.viewInstanceHandler)

	go func() {
		for {
			s.CheckHeartbeats()
			<-time.After(1 * time.Second)
		}
	}()

	log.Printf("listening to %s", s.ListenAddr)
	return http.ListenAndServe(s.ListenAddr, router)
}

// GetApplication return the Application which has the coresponding name.
// Return nil if no Application with this name has been found.
func (s *Server) GetApplication(name string) *Application {
	for _, app := range s.Applications {
		if app.Name == name {
			return app
		}
	}
	return nil
}

// CheckHeartbeats update Applications status depending on received heartbeats.
// It may also remove unresponsive instances.
func (s *Server) CheckHeartbeats() {
	for _, app := range s.Applications {
		app.CheckHeartbeats()
	}
}

// listAppsHandler is the HTTP handler for /apps
func (s *Server) listAppsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// List all applications registered to the server
		listApps(s.Applications, w, r)
	case "POST":
		// Register a new application
		app, err := newApp(w, r)
		if err != nil {
			log.Printf("%s", err)
			return
		}

		// Check if app already exists
		if s.GetApplication(app.Name) != nil {
			w.WriteHeader(409)
			return
		}

		// Add application
		s.Applications = append(s.Applications, app)
		w.WriteHeader(201)
		log.Printf("new application created: %s", app.Name)
	default:
		// Unsuported method
		w.WriteHeader(405)
	}
}

// listApps writes the list of applications to w.
func listApps(apps []*Application, w http.ResponseWriter, r *http.Request) {
	var response struct {
		Apps []*Application `json:"applications"`
	}
	response.Apps = make([]*Application, 0)
	for _, app := range apps {
		response.Apps = append(response.Apps, app)
	}

	// Marshal and write response
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	fmt.Fprintln(w, string(data))
}

// newApp return a new application from the r.Body.
func newApp(w http.ResponseWriter, r *http.Request) (*Application, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return nil, err
	}

	// Unmarshal request and return
	var request struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		w.WriteHeader(400)
		return nil, err
	}
	app := NewApplication(request.Name)
	return app, nil
}

// viewAppHandler is the HTTP handler for /apps/{appName}.
func (s *Server) viewAppHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	app := s.GetApplication(vars["appName"])
	if app == nil {
		w.WriteHeader(404)
		return
	}

	switch r.Method {
	case "GET":
		// Show app details
		viewApp(app, w, r)
	case "POST":
		// New app instance
		inst, err := newInstance(w, r)
		if err != nil {
			log.Printf("%s", err)
			return
		}

		// Check if instance already exists
		if app.GetInstance(inst.Id) != nil {
			w.WriteHeader(409)
			return
		}

		// Add instance
		app.Instances = append(app.Instances, inst)
		w.WriteHeader(201)
		log.Printf("instance %s added to app %s", inst.Id, app.Name)
	}
}

// viewApp writes the app details to w.
func viewApp(app *Application, w http.ResponseWriter, r *http.Request) {
	data, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	fmt.Fprintln(w, string(data))
}

// newInstance return a new application instance from r.Body.
func newInstance(w http.ResponseWriter, r *http.Request) (*Instance, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return nil, err
	}

	var request struct {
		Id   string `json:"id"`
		Ip   string `json:"ip"`
		Port int    `json:"port"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		w.WriteHeader(400)
		return nil, err
	}

	// Check if everything is set
	if request.Id == "" || request.Ip == "" || request.Port == 0 {
		w.WriteHeader(400)
		return nil, errors.New("required parameter missing")
	}

	inst := NewInstance(request.Id, request.Ip, request.Port)
	return inst, nil
}

// viewInstanceHandler is the HTTP handler for /apps/{appName}/{instanceId}.
func (s *Server) viewInstanceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	app := s.GetApplication(vars["appName"])
	if app == nil {
		w.WriteHeader(404)
		return
	}

	inst := app.GetInstance(vars["instanceId"])
	if inst == nil {
		w.WriteHeader(404)
		return
	}

	switch r.Method {
	case "GET":
		// Show instance details
		viewInstance(inst, w, r)
	case "PUT":
		// Renew instance heartbeat
		renewInstance(inst, w, r)
	case "DELETE":
		// Put instance out-of-service
		deleteInstance(inst, w, r)
		log.Printf("instance %s is out-of-service", inst.Id)
	}
}

// viewInstance writes the instance details to w.
func viewInstance(inst *Instance, w http.ResponseWriter, r *http.Request) {
	data, err := json.MarshalIndent(inst, "", "  ")
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	fmt.Fprintln(w, string(data))
}

// renewInstance updates the instance heartbeat.
// It also changes the status to UP.
func renewInstance(inst *Instance, w http.ResponseWriter, r *http.Request) {
	if inst.Status == OUTOFSERVICE {
		log.Printf("cannot renew out-of-service instance %s", inst.Id)
		w.WriteHeader(403)
		return
	}

	if inst.Status != UP {
		log.Printf("instance %s is now UP", inst.Id)
		inst.Status = UP
	}
	inst.Touch()
	w.WriteHeader(204)
}

// deleteInstance put an instance out-of-order.
// If an instance is out-of-service it cannot be restarted and may be deleted after a time.
func deleteInstance(inst *Instance, w http.ResponseWriter, r *http.Request) {
	inst.Status = OUTOFSERVICE
	inst.Touch()
	w.WriteHeader(204)
}
