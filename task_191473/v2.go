//go:build v2
// +build v2

package main

// Importing necessary packages
import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"embed"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"

	_ "github.com/mattn/go-sqlite3"
)

// Embedding the database schema file.
//go:embed schema/tagsdb.sql
var schemaFS embed.FS

// Configuration structure and related functions
type Config struct {
	CertFile                string `mapstructure:"cert_file" json:"cert_file"`
	DatabasePath            string `mapstructure:"database_path" json:"database_path"`
	KeyFile                 string `mapstructure:"key_file" json:"key_file"`
	TagIdFieldName          string `mapstructure:"tag_id_field_name" json:"tag_id_field_name"`
	TrainingFieldName       string `mapstructure:"training_field_name" json:"training_field_name"`
	WildApricotAccountId    int    `mapstructure:"wild_apricot_account_id" json:"wild_apricot_account_id"`
	ContactFilterQuery      string `mapstructure:"contact_filter_query" json:"contact_filter_query"`
	SSOClientID             string `mapstructure:"sso_client_id" json:"sso_client_id"`
	SSOClientSecret         string `mapstructure:"sso_client_secret" json:"sso_client_secret"`
	SSORedirectURI          string `mapstructure:"sso_redirect_uri" json:"sso_redirect_uri"`
	CookieStoreSecret       string `mapstructure:"cookie_store_secret" json:"cookie_store_secret"`
	WildApricotApiKey       string
	WildApricotWebhookToken string
	LogDir                  string `mapstructure:"log_dir" json:"log_dir"`
	log                     *logrus.Logger
}

var (
	config *utils.Singleton
	once   sync.Once
)

func loadConfig() interface{} {
	projectRoot, err := utils.GetProjectRoot()
	if err != nil {
		log.Fatalf("Error fetching project root absolute path: %s", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(projectRoot)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Error unmarshalling config file: %s", err)
	}

	cfg.CertFile = filepath.Join(projectRoot, cfg.CertFile)
	if _, err := os.Stat(cfg.CertFile); os.IsNotExist(err) {
		log.Fatalf("Certificate file not found: %s", cfg.CertFile)
	}

	cfg.KeyFile = filepath.Join(projectRoot, cfg.KeyFile)
	if _, err := os.Stat(cfg.KeyFile); os.IsNotExist(err) {
		log.Fatalf("Key file not found: %s", cfg.KeyFile)
	}

	// Validate and set log directory
	cfg.LogDir = filepath.Join(projectRoot, cfg.LogDir)
	if err := createLogDir(cfg.LogDir); err != nil {
		log.Fatalf("Error creating log directory: %s", err)
	}

	// Load environment variables
	cfg.WildApricotApiKey = os.Getenv("WILD_APRICOT_API_KEY")
	if cfg.WildApricotApiKey == "" {
		log.Fatalf("WILD_APRICOT_API_KEY not set in environment variables")
	}

	cfg.WildApricotWebhookToken = os.Getenv("WILD_APRICOT_WEBHOOK_TOKEN")
	if cfg.WildApricotWebhookToken == "" {
		log.Fatalf("WILD_APRICOT_WEBHOOK_TOKEN not set in environment variables")
	}

	cfg.SSOClientID = os.Getenv("WILD_APRICOT_SSO_CLIENT_ID")
	if cfg.SSOClientID == "" {
		log.Fatalf("WILD_APRICOT_SSO_CLIENT_ID not set in environment variables")
	}

	cfg.SSOClientSecret = os.Getenv("WILD_APRICOT_SSO_CLIENT_SECRET")
	if cfg.SSOClientSecret == "" {
		log.Fatalf("WILD_APRICOT_SSO_CLIENT_SECRET not set in environment variables")
	}

	cfg.SSORedirectURI = os.Getenv("WILD_APRICOT_SSO_REDIRECT_URI")
	if cfg.SSORedirectURI == "" {
		log.Fatalf("WILD_APRICOT_SSO_REDIRECT_URI not set in environment variables")
	}

	return &cfg
}

func LoadConfig() *Config {
	return config.Get(loadConfig).(*Config)
}

func createLogDir(dir string) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// Setup Logger
func SetupLogger(cfg *Config) *logrus.Logger {
	logger := logrus.New()

	// Log to file
	logFilePath := filepath.Join(cfg.LogDir, "app.log")
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %s", err)
	}

	multiWriter := io.MultiWriter(file, os.Stdout)
	logger.SetOutput(multiWriter)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	logLevel, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	return logger
}

// Utility functions
package utils

import (
"os"
"path/filepath"
"strings"
"sync"
)

func GetProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	projectRootMarker := "rfid-backend"

	if strings.Contains(currentDir, projectRootMarker) {
		index := strings.Index(currentDir, projectRootMarker)

		projectRoot := currentDir[:index+len(projectRootMarker)]

		return filepath.Clean(projectRoot), nil
	}

	return filepath.Clean(currentDir), nil
}

type Singleton struct {
	once sync.Once
	obj  interface{}
}

func NewSingleton(obj interface{}) *Singleton {
	return &Singleton{obj: obj}
}

func (s *Singleton) Get(initialize func() interface{}) interface{} {
	s.once.Do(func() {
		s.obj = initialize()
	})
	return s.obj
}

// Models
package models

type MemberTrainingLink struct {
	TagID        uint32 // Foreign Key to Members (is an RFID)
	TrainingName string // Foreign Key to Trainings
}

type Training struct {
	TrainingName string
}

type Device struct {
	IPAddress        string
	MACAddress       string
	RequiresTraining int
}

type DeviceTrainingLink struct {
	MACAddress string
	Label      string
}

type Member struct {
	TagId           uint32 // corresponds to RFIDFieldName in config
	MembershipLevel int
}

type Contact struct {
	FirstName              string       `json:"FirstName"`
	LastName               string       `json:"LastName"`
	Email                  string       `json:"Email"`
	DisplayName            string       `json:"DisplayName"`
	Organization           string       `json:"Organization"`
	ProfileLastUpdated     string       `json:"ProfileLastUpdated"`
	FieldValues            []FieldValue `json:"FieldValues"`
	Id                     int          `json:"Id"`
	Url                    string       `json:"Url"`
	IsAccountAdministrator bool         `json:"IsAccountAdministrator"`
	TermsOfUseAccepted     bool         `json:"TermsOfUseAccepted"`
	Status                 string       `json:"Status"`
}

type FieldValue struct {
	FieldName  string      `json:"FieldName"`
	Value      interface{} `json:"Value"`
	SystemCode string      `json:"SystemCode"`
}

type SafetyTraining struct {
	Id    int    `json:"Id"`
	Label string `json:"Label"`
}

func (c *Contact) ExtractTagID(cfg *Config) (uint32, error) {
	for _, val := range c.FieldValues {
		if val.FieldName == cfg.TagIdFieldName {
			return parseTagId(val)
		}
	}
	return 0, nil // Return 0 if TagId field is not found
}

func (c *Contact) ExtractTrainingLabels(cfg *Config) ([]string, error) {
	for _, val := range c.FieldValues {
		if val.FieldName == cfg.TrainingFieldName {
			return parseTrainingLabels(val)
		}
	}
	return nil, nil // Return nil if Training field is not found
}

func (c *Contact) ExtractContactData(cfg *Config) (int, uint32, []string, error) {
	tagID, err := c.ExtractTagID(cfg)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("error extracting TagId for contact %d: %v", c.Id, err)
	}

	trainingLabels, err := c.ExtractTrainingLabels(cfg)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("error extracting training labels for contact %d: %v", c.Id, err)
	}

	return c.Id, tagID, trainingLabels, nil
}

// Database operations
package db

import (
"database/sql"
"fmt"

"github.com/dlclark/regexp2"
)

type Database struct {
	Db *sql.DB
}

func (db *Database) CreateTables() error {
	// Read the schema from the embedded file system
	schema, err := schemaFS.ReadFile("schema/tagsdb.sql")
	if err != nil {
		return fmt.Errorf("error reading schema: %v", err)
	}

	// Execute the SQL commands to create tables
	_, err = db.Db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("error executing schema: %v", err)
	}
	return nil
}

// Middleware
package middleware

import (
"github.com/gin-gonic/gin"
"net/http"
)

func AuthRequired(c *gin.Context) {
	// Your authentication logic
	if !isAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}
	c.Next()
}

func isAuthenticated(c *gin.Context) bool {
	// Implement your authentication check here
	return true // This is a placeholder
}

// Main function
func main() {
	cfg := LoadConfig()

	// Setup logger
	cfg.log = SetupLogger(cfg)

	// Database initialization
	db, err := sql.Open("sqlite3", cfg.DatabasePath)
	if err != nil {
		cfg.log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	database := &db.Database{Db: db}
	if err := database.CreateTables(); err != nil {
		cfg.log.Fatalf("Error creating tables: %v", err)
	}

	// Setting up Gin router
	r := gin.Default()
	r.Use(sessions.Sessions("mysession", cookie.NewStore([]byte(cfg.CookieStoreSecret))))
	r.Use(middleware.AuthRequired)

	// Your routes here
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to the access control system!"})
	})

	// Start the server
	if err := r.RunTLS(":443", cfg.CertFile, cfg.KeyFile); err != nil {
		cfg.log.Fatalf("Error starting server: %v", err)
	}
}