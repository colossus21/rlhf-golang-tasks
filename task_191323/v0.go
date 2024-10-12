//go:build v0
// +build v0

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Constants
const DOCKER_JWT_STOPPED = "Docker rest_api_jwt_pdv_app com problema"
const DOCKER_JWT_RUNNING = "Docker rest_api_jwt_pdv_app rodando normalmente"
const TIME_DELAY = 15

// Docker Model
type Docker struct {
	Name      string
	Running   bool
	CheckedAt time.Time
}

// Server Struct
type Server struct {
	Router *mux.Router
}

// Middleware
func SetMiddlewareJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	}
}

// Response Functions
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func ERROR(w http.ResponseWriter, statusCode int, err error) {
	if err != nil {
		JSON(w, statusCode, struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
		return
	}
	JSON(w, http.StatusBadRequest, nil)
}

// Services Function
func GetAllDockers() []Docker {
	c := &http.Client{
		Timeout: TIME_DELAY * time.Second,
	}

	time.Local, _ = time.LoadLocation("America/Sao_Paulo")

	var dockers []Docker

	for i := 0; i < 3; i++ {
		req, err := c.Get(os.Getenv("DOCKER_API_JWT"))
		docker := new(Docker)
		if err != nil {
			docker.Name = DOCKER_JWT_STOPPED
			docker.Running = false
			docker.CheckedAt = time.Now()
		} else {
			_, err := ioutil.ReadAll(req.Body)
			if err != nil {
				docker.Name = DOCKER_JWT_STOPPED
				docker.Running = false
				docker.CheckedAt = time.Now()
			} else {
				docker.Name = DOCKER_JWT_RUNNING
				docker.Running = true
				docker.CheckedAt = time.Now()
			}
		}
		dockers = append(dockers, *docker)
	}

	return dockers
}

// Initialize and Run Server
func (server *Server) Initialize() {
	server.Router = mux.NewRouter()
	server.initializeRoutes()
}

func (server *Server) Run(addr string) {
	fmt.Println("Listening to port 8295")
	log.Fatal(http.ListenAndServe(addr, server.Router))
}

// Initialize Routes
func (s *Server) initializeRoutes() {
	s.Router.HandleFunc("/", SetMiddlewareJSON(s.GetAllDockers)).Methods("GET")
}

// Controller Function
func (server *Server) GetAllDockers(w http.ResponseWriter, r *http.Request) {
	resp := GetAllDockers()
	JSON(w, http.StatusOK, resp)
}

// Entry point
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, not coming through %v", err)
	}

	server := Server{}
	server.Initialize()
	server.Run(":8295")
}
