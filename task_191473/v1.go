//go:build v1
// +build v1

package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"

	_ "github.com/mattn/go-sqlite3"
	"rfid-backend/config"   // Assuming utils is in a package named config
	"rfid-backend/models"
	"rfid-backend/services"
	"rfid-backend/setup"    // Assuming setup functions are in a setup package
	"rfid-backend/utils"
)

//go:embed schema/tagsdb.sql
var schemaFS embed.FS

// Config structure and related functions
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
	LogDir                  string `mapstructure:"log_dir" json:"log_dir"` // New field for log directory
	log                     *logrus.Logger
}

var (
	configInstance *utils.Singleton
	once           sync.Once
)

// CreateLogDir creates the log directory if it doesn't exist.
func CreateLogDir(cfg *Config) error {
	logDir := cfg.LogDir
	if logDir == "" {
		return nil // No custom log directory specified, use default
	}

	err := os.MkdirAll(logDir, 0755) // Create directory with read/write/execute permissions for owner and read/execute for others
	if err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	return nil
}

// Updated SetupLogger function
func SetupLogger(cfg *Config) *logrus.Logger {
	logger := logrus.New()

	// Setup timestamp format
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339, // Or any other desired format
	})

	if cfg.LogDir != "" {
		logFilePath := filepath.Join(cfg.LogDir, "rfid-backend.log")
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			logger.Fatalf("Failed to open log file: %v", err) // Fatal error if log file can't be opened
		}
		logger.SetOutput(file)
	}

	logLevel, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	return logger
}

// LoadConfig function
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

		return &cfg
	}

	// LoadConfig returns a singleton instance of Config.
	func
	LoadConfig() * Config{
		return configInstance.Get(loadConfig).(*Config)
	}

	func
	main()
	{
		fmt.Println("Starting RFID Backend - Access Control System")

		cfg := LoadConfig()

		// Creating log directory before initializing logger
		if err := CreateLogDir(cfg); err != nil {
			log.Fatalf("Failed to create config directory: %s", err) // Use standard log for errors during early initialization
		}
		logger := setup.SetupLogger(cfg) // Pass config to SetupLogger

		db, err := setup.SetupDatabase(cfg, logger)
		if err != nil {
			logger.Fatalf("Failed to setup database: %v", err)
		}
		defer db.Close()

		// Initialize singleton using the newly created logger instance
		config.Init(cfg, logger)

		router := gin.Default()
		setup.SetupRoutes(router, cfg, db, logger)

		waService := services.NewWildApricotService(config.GetConfig(), logger)
		dbService := services.NewDBService(db, config.GetConfig(), logger)

		services.StartBackgroundDatabaseUpdate(waService, dbService, logger)

		if err := router.RunTLS(":443", cfg.CertFile, cfg.KeyFile); err != nil {
			logger.Fatalf("Failed to start HTTPS server: %v", err)
		}
	}
}