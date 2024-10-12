//go:build v1
// +build v1

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Constants
const DOCKER_JWT_STOPPED = "Docker rest_api_jwt_pdv_app com problema"
const DOCKER_JWT_RUNNING = "Docker rest_api_jwt_pdv_app rodando normalmente"

// Docker Model
type Docker struct {
	Name         string        `json:"name"`
	Running      bool          `json:"running"`
	CheckedAt    time.Time     `json:"checkedAt"`
	StatusCode   int           `json:"statusCode"`
	ResponseTime time.Duration `json:"responseTime"`
}

// Server Struct
type Server struct {
	Router *mux.Router
	Client *http.Client
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
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Services Function
func GetAllDockers(client *http.Client, dockerAPIURL string, timeDelay time.Duration) []Docker {
	var dockers []Docker
	var wg sync.WaitGroup
	var mutex sync.Mutex // Protect dockers slice from concurrent writes

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			req, err := client.Get(dockerAPIURL)
			elapsed := time.Since(start)

			docker := Docker{
				CheckedAt:    time.Now(),
				ResponseTime: elapsed,
			}

			if err != nil {
				log.Printf("Error making request: %v", err)
				docker.Name = DOCKER_JWT_STOPPED
				docker.Running = false
				docker.StatusCode = -1 // Or another appropriate error code

			} else {
				defer req.Body.Close()
				body, err := ioutil.ReadAll(req.Body)
				if err != nil {
					log.Printf("Error reading response body: %v", err)
					docker.Name = DOCKER_JWT_STOPPED
					docker.Running = false
					docker.StatusCode = req.StatusCode
				} else {

					docker.StatusCode = req.StatusCode
					if req.StatusCode == http.StatusOK {
						docker.Name = DOCKER_JWT_RUNNING
						docker.Running = true
						log.Printf("Docker check successful. Status code: %d, Response Body: %s", req.StatusCode, string(body))

					} else {
						log.Printf("Docker check failed. Status code: %d, Response Body: %s", req.StatusCode, string(body))

						docker.Name = DOCKER_JWT_STOPPED
						docker.Running = false
					}
				}

			}

			mutex.Lock()
			dockers = append(dockers, docker)
			mutex.Unlock()
		}()
	}

	wg.Wait()
	return dockers
}

// Initialize and Run Server
func (server *Server) Initialize() {
	timeDelayStr := os.Getenv("TIME_DELAY")
	timeDelay, err := strconv.Atoi(timeDelayStr)
	if err != nil {
		log.Printf("Invalid TIME_DELAY value: %s. Using default value of 15.", timeDelayStr)

		timeDelay = 15
	}

	dockerAPIURL := os.Getenv("DOCKER_API_JWT")
	if dockerAPIURL == "" {
		log.Fatal("DOCKER_API_JWT environment variable not set")
	}

	server.Client = &http.Client{
		Timeout: time.Duration(timeDelay) * time.Second,
	}

	server.Router = mux.NewRouter()
	server.initializeRoutes(dockerAPIURL, time.Duration(timeDelay)*time.Second)
}

func (server *Server) Run(addr string) {
	log.Printf("Listening to port %s", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: server.Router,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
		log.Println("Server gracefully stopped")
	}()

	log.Fatal(srv.ListenAndServe())

}

// Initialize Routes
func (s *Server) initializeRoutes(dockerAPIURL string, timeDelay time.Duration) {
	s.Router.HandleFunc("/", SetMiddlewareJSON(func(w http.ResponseWriter, r *http.Request) {
		s.GetAllDockers(w, r, dockerAPIURL, timeDelay)
	})).Methods("GET")
}

// Controller Function
func (server *Server) GetAllDockers(w http.ResponseWriter, r *http.Request, dockerAPIURL string, timeDelay time.Duration) {
	resp := GetAllDockers(server.Client, dockerAPIURL, timeDelay)
	JSON(w, http.StatusOK, resp)

}

// Entry point
func main() {
	time.Local, _ = time.LoadLocation("America/Sao_Paulo") // Correct

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	server := Server{}
	server.Initialize()
	server.Run(":8295")
}
