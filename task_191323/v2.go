//go:build v2
// +build v2

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Constants
const (
	dockerJWTStoppedMsg = "Docker rest_api_jwt_pdv_app com problema"
	dockerJWTRunningMsg = "Docker rest_api_jwt_pdv_app rodando normalmente"
	defaultTimeDelay    = 15 // Default delay in seconds
)

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
	Client *http.Client // HTTP client for reuse
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
func GetAllDockers(client *http.Client) []Docker {
	timeDelay := getEnvInt("TIME_DELAY", defaultTimeDelay)

	time.Local, _ = time.LoadLocation("America/Sao_Paulo")

	dockerAPIJWT := os.Getenv("DOCKER_API_JWT")
	if dockerAPIJWT == "" {
		log.Fatal("DOCKER_API_JWT environment variable not set")
	}

	var dockers []Docker
	var wg sync.WaitGroup
	dockerChan := make(chan Docker, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			startTime := time.Now()
			req, err := client.Get(dockerAPIJWT)
			elapsedTime := time.Since(startTime)

			docker := Docker{
				CheckedAt:    time.Now(),
				ResponseTime: elapsedTime,
			}

			if err != nil {
				log.Printf("Error checking Docker: %s", err.Error())
				docker.Name = dockerJWTStoppedMsg
				docker.Running = false
			} else {
				defer req.Body.Close()
				_, err := ioutil.ReadAll(req.Body)
				if err != nil {
					log.Printf("Error reading response body: %s", err.Error())
					docker.Name = dockerJWTStoppedMsg
					docker.Running = false
				} else {
					docker.Name = dockerJWTRunningMsg
					docker.Running = true
					docker.StatusCode = req.StatusCode
					log.Printf("Docker check successful. Status Code: %d", req.StatusCode)
				}
			}
			dockerChan <- docker
		}()
	}

	go func() {
		wg.Wait()
		close(dockerChan)
	}()

	for d := range dockerChan {
		dockers = append(dockers, d)
	}

	time.Sleep(time.Duration(timeDelay) * time.Second)
	return dockers
}

// Initialize and Run Server
func (server *Server) Initialize() {
	server.Router = mux.NewRouter()
	server.Client = &http.Client{Timeout: time.Second * 30}
	server.initializeRoutes()
}

func (server *Server) Run(addr string) {
	fmt.Println("Listening to port 8295")

	// Create a server with graceful shutdown
	srv := &http.Server{
		Addr:    addr,
		Handler: server.Router,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Println("Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped gracefully")
}

// Initialize Routes
func (s *Server) initializeRoutes() {
	s.Router.HandleFunc("/", SetMiddlewareJSON(s.GetAllDockers)).Methods("GET")
}

// Controller Function
func (server *Server) GetAllDockers(w http.ResponseWriter, r *http.Request) {
	resp := GetAllDockers(server.Client)
	JSON(w, http.StatusOK, resp)
}

// Helper function to get environment variable as int with default value
func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Invalid value for environment variable %s: %s. Using default value %d.", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
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
