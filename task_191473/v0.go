//go:build v0
// +build v0

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

func LoadConfig() *Config {
	return config.Get(loadConfig).(*Config)
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
	TagID        uint32 // Foreign Key to Members (is an rfid)
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

func (c *Contact) ExtractTagID(cfg *config.Config) (uint32, error) {
	for _, val := range c.FieldValues {
		if val.FieldName == cfg.TagIdFieldName {
			return parseTagId(val)
		}
	}
	return 0, nil // Return 0 if TagId field is not found
}

func (c *Contact) ExtractTrainingLabels(cfg *config.Config) ([]string, error) {
	for _, val := range c.FieldValues {
		if val.FieldName == cfg.TrainingFieldName {
			return parseTrainingLabels(val)
		}
	}
	return nil, nil // Return nil if Training field is not found
}

func (c *Contact) ExtractContactData(cfg *config.Config) (int, uint32, []string, error) {
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

func parseTagId(fieldValue FieldValue) (uint32, error) {
	strVal, ok := fieldValue.Value.(string)
	if !ok {
		return 0, errors.New("TagId value is not a string")
	}

	if len(strVal) <= 0 {
		// Suppress error on empty TagId field value, return 0
		return uint32(0), nil
	}

	tagId, err := strconv.ParseInt(strVal, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string TagId to int: %v", err)
	}

	if tagId <= 0 {
		return 0, errors.New("TagId value is non-positive")
	}

	return uint32(tagId), nil
}

func parseTrainingLabels(fieldValue FieldValue) ([]string, error) {
	trainingValues, ok := fieldValue.Value.([]interface{})
	if !ok {
		return nil, errors.New("training value is not a slice")
	}

	var trainings []string
	for _, item := range trainingValues {
		training, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unable to parse training item")
		}
		if label, ok := training["Label"].(string); ok {
			trainings = append(trainings, label)
		}
	}
	return trainings, nil
}

// Database initialization
func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	schemaFile, err := fs.ReadFile(schemaFS, "schema/tagsdb.sql")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(string(schemaFile))
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Services
package services

type DBService struct {
	db  *sql.DB
	cfg *config.Config
	log *logrus.Logger
}

func NewDBService(db *sql.DB, cfg *config.Config, logger *logrus.Logger) *DBService {
	return &DBService{db: db, cfg: cfg, log: logger}
}

const (
	TagExistsQuery = `
SELECT EXISTS(SELECT 1 FROM members WHERE tag_id = ?)
`
	GetTagIdsForTrainingQuery = `
SELECT tag_id
FROM members_trainings_link
WHERE label = ?;
`
	GetTrainingQuery = `
SELECT label
FROM trainings
WHERE label = ?
`
	GetAllTrainingsQuery = `
SELECT label
FROM trainings
`
	GetAllDevicesQuery = `
SELECT ip_address, mac_address
FROM devices;
`
	GetAllTagIdsQuery = `
SELECT tag_id
FROM members;
`
	GetAllDevicesTrainingsQuery = `
SELECT mac_address, label
FROM devices_trainings_link;
`
	InsertOrUpdateMemberQuery = `
INSERT OR IGNORE INTO members (contact_id, tag_id, membership_level)
VALUES (?, ?, ?)
ON CONFLICT(contact_id) DO UPDATE SET tag_id = ?, membership_level = EXCLUDED.membership_level;
`
	InsertTrainingQuery = `
INSERT OR IGNORE INTO trainings (label)
VALUES (?);
`
	InsertMemberTrainingLinkQuery = `
INSERT OR IGNORE INTO members_trainings_link (tag_id, label)
VALUES (?, ?);
`
	InsertDeviceQuery = `
INSERT OR IGNORE INTO devices (ip_address, mac_address, requires_training)
VALUES (?, ?, ?);
`
	InsertDeviceTrainingLinkQuery = `
INSERT INTO devices_trainings_link (mac_address, label)
VALUES (?, ?)
ON CONFLICT(mac_address, label) DO UPDATE
SET mac_address = EXCLUDED.mac_address;
`
	DeleteInactiveMembersQuery = `
DELETE FROM members WHERE contact_id NOT IN (%s)
`
	DeleteLapsedMembersQuery = `
DELETE FROM members WHERE contact_id = %s
`
	DeleteDeviceTrainingLinkQuery = `
DELETE FROM devices_trainings_link
WHERE mac_address = ?;
`
)

func (s *DBService) GetAllTagIds() ([]uint32, error) {
	return s.fetchTagIds(GetAllTagIdsQuery)
}

func (s *DBService) fetchTagIds(query string, args ...interface{}) ([]uint32, error) {
	var tagIds []uint32

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tagId uint32
		if err := rows.Scan(&tagId); err != nil {
			return nil, err
		}
		tagIds = append(tagIds, tagId)
	}

	return tagIds, nil
}

func (s *DBService) GetTraining(label string) (string, error) {
	row, err := s.db.Query(GetTrainingQuery, label)
	if err != nil {
		return "", err
	}

	var training string
	if err := row.Scan(&training); err != nil {
		s.log.Warnf("No Training found for %s", label)
		return "", err
	}

	return training, nil
}

func (s *DBService) GetAllTrainings() ([]string, error) {
	var devices []string

	rows, err := s.db.Query(GetAllTrainingsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}

	return devices, nil
}

type WildApricotService struct {
	Client             *http.Client
	cfg                *config.Config
	TokenEndpoint      string
	ApiToken           string
	WildApricotApiBase string
	TokenExpiry        time.Time
	log                *logrus.Logger
}

var wildApricotSvc = utils.NewSingleton(&WildApricotService{})

func NewWildApricotService(cfg *config.Config, logger *logrus.Logger) *WildApricotService {
	return wildApricotSvc.Get(func() interface{} {
		s := &WildApricotService{
			Client: &http.Client{
				Timeout: time.Second * 30,
			},
			cfg:                cfg,
			TokenEndpoint:      "https://oauth.wildapricot.org/auth/token",
			WildApricotApiBase: "https://api.wildapricot.org/v2/accounts",
			log:                logger,
		}
		s.log.Info("WildApricotService initialized")
		return s
	}).(*WildApricotService)
}

func readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func handleHTTPError(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return nil
}

func (s *WildApricotService) logError(context string, err error) {
	if err != nil {
		s.log.WithFields(logrus.Fields{"context": context, "error": err}).Error("Error occurred")
	}
}

func (s *WildApricotService) buildURL(pathFormat string, args ...interface{}) string {
	return fmt.Sprintf(s.WildApricotApiBase+pathFormat, args...)
}

func unmarshalJSON(body []byte, target interface{}) error {
	return json.Unmarshal(body, target)
}

func (s *WildApricotService) refreshTokenIfNeeded() error {
	if time.Now().After(s.TokenExpiry) || s.ApiToken == "" {
		s.log.Info("Refreshing API token")
		return s.refreshApiToken()
	}
	return nil
}

func (s *WildApricotService) refreshApiToken() error {
	url := s.TokenEndpoint
	data := "grant_type=client_credentials&scope=auto"
	encodedApiKey := base64.StdEncoding.EncodeToString([]byte("APIKEY:" + s.cfg.WildApricotApiKey))
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		s.logError("Error creating token refresh request: %v", err)
		return err
	}
	req.Header.Add("Authorization", "Basic "+encodedApiKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		s.logError("Error during token refresh: %v", err)
		return err
	}

	body, err := readResponseBody(resp)

	if err != nil {
		s.logError("Error reading token refresh response: %v", err)
		return err
	}

	if err := handleHTTPError(resp); err != nil {
		s.logError("HTTP error during token refresh: %v", err)
		return err
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := unmarshalJSON(body, &tokenResp); err != nil {
		s.logError("Error unmarshalling token response: %v", err)
		return err
	}

	s.ApiToken = tokenResp.AccessToken
	s.TokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	s.log.Info("API token refreshed successfully")

	return nil
}

func (s *WildApricotService) GetContacts() ([]models.Contact, error) {
	if err := s.refreshTokenIfNeeded(); err != nil {
		return nil, err
	}

	url := s.buildURL("/%d/Contacts?$filter=%s", s.cfg.WildApricotAccountId, url.QueryEscape(s.cfg.ContactFilterQuery))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.logError("Error creating contact request: %v", err)
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+s.ApiToken)

	resp, err := s.Client.Do(req)
	if err != nil {
		s.logError("Error fetching contacts: %v", err)
		return nil, err
	}

	body, err := readResponseBody(resp)
	if err != nil {
		s.logError("Error reading contact response: %v", err)
		return nil, err
	}

	if err := handleHTTPError(resp); err != nil {
		s.logError("HTTP error fetching contacts: %v", err)
		return nil, err
	}

	var contacts []models.Contact
	if err := unmarshalJSON(body, &contacts); err != nil {
		s.logError("Error unmarshalling contact response: %v", err)
		return nil, err
	}

	return contacts, nil
}

func (s *DBService) ProcessContactsData(contacts []models.Contact) error {
	var err error

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, contact := range contacts {
		contactId, tagId, trainingLabels, err := contact.ExtractContactData(s.cfg)
		if err != nil {
			s.log.Warnf("Failed to extract data for contact: %v", err)
			continue
		}

		if tagId > 0 {
			err = s.InsertOrUpdateMember(tx, contactId, tagId, contact.MembershipLevel)
			if err != nil {
				return fmt.Errorf("failed to insert or update member: %w", err)
			}

			for _, trainingLabel := range trainingLabels {
				if err := s.InsertTraining(tx, trainingLabel); err != nil {
					return fmt.Errorf("failed to insert training: %w", err)
				}

				if err := s.InsertMemberTrainingLink(tx, tagId, trainingLabel); err != nil {
					return fmt.Errorf("failed to insert member training link: %w", err)
				}
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *DBService) InsertOrUpdateMember(tx *sql.Tx, contactId int, tagId uint32, membershipLevel int) error {
	_, err := tx.Exec(InsertOrUpdateMemberQuery, contactId, tagId, membershipLevel, tagId)
	return err
}

func (s *DBService) InsertTraining(tx *sql.Tx, label string) error {
	_, err := tx.Exec(InsertTrainingQuery, label)
	return err
}

func (s *DBService) InsertMemberTrainingLink(tx *sql.Tx, tagId uint32, label string) error {
	_, err := tx.Exec(InsertMemberTrainingLinkQuery, tagId, label)
	return err
}

// Main setup and configuration functions
// Setup Logger
package setup

func SetupLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	logLevel, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	return logger
}

// Setup Database
func SetupDatabase(cfg *config.Config, logger *logrus.Logger) (*sql.DB, error) {
	database, err := InitDB(cfg.DatabasePath)
	if err != nil {
		logger.Errorf("Failed to initialize database: %v", err)
		return nil, err
	}
	return database, nil
}

// Setup Routes
func SetupRoutes(router *gin.Engine, cfg *config.Config, db *sql.DB, logger *logrus.Logger) {
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	waService := NewWildApricotService(cfg, logger)
	dbService := NewDBService(db, cfg, logger)

	oauthConf := &oauth2.Config{
		ClientID:     cfg.SSOClientID,
		ClientSecret: cfg.SSOClientSecret,
		RedirectURL:  cfg.SSORedirectURI,
		Scopes:       []string{"contacts_me"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://davidbouw36728.wildapricot.org/sys/login/OAuthLogin",
			TokenURL: "https://oauth.wildapricot.org/auth/token",
		},
	}
	Initialize(oauthConf, cfg, logger)

	authGroup := router.Group("/auth")
	{
		authGroup.GET("/login", StartOAuthFlow)
		authGroup.GET("/callback", OAuthCallback)
	}

	registrationHandler := NewRegistrationHandler(dbService, cfg, logger)
	api := router.Group("/api")
	{
		webhooksHandler := NewWebhooksHandler(waService, dbService, cfg, logger)
		configHandler := NewConfigHandler(logger)
		accessControlHandler := NewAccessControlHandler(dbService, logger)

		api.POST("authenticate", accessControlHandler.HandleAuthenticate)
		api.POST("/updateConfig", configHandler.UpdateConfig)
		api.POST("/webhooks", webhooksHandler.HandleWebhook)
		api.POST("/register", registrationHandler.HandleRegisterDevice)
		api.POST("/updateDeviceAssignments", registrationHandler.UpdateDeviceAssignments)
	}

	router.Static("/css", "./web-ui/css")
	router.Static("/js", "./web-ui/js")
	router.Static("/assets", "./web-ui/assets")
	router.LoadHTMLGlob("web-ui/templates/*")

	setupWebUIRoutes(router, dbService, cfg, logger)
}

func setupWebUIRoutes(router *gin.Engine, dbService *NewDBService(db, cfg, logger), cfg *config.Config, logger *logrus.Logger) {
	rh := NewRegistrationHandler(dbService, cfg, logger)
	webUI := router.Group("/web-ui")
	{
		webUI.Use(RequireAuth)
		webUI.GET("/home", func(c *gin.Context) {
			logger.Info("Serving the home page")
			c.HTML(http.StatusOK, "home.tmpl", nil)
		})
		webUI.GET("/configManagement", func(c *gin.Context) {
			c.HTML(http.StatusOK, "configManagement.tmpl", gin.H{"title": "Configuration Management"})
		})
		webUI.GET("/deviceManagement", rh.ServeDeviceManagementPage)
	}
}

// Main function
func main() {
	fmt.Println("Starting RFID Backend - Access Control System")

	logger := SetupLogger()

	cfg := LoadConfig()
	db, err := SetupDatabase(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to setup database: %v", err)
	}
	defer db.Close()

	router := gin.Default()
	SetupRoutes(router, cfg, db, logger)

	waService := NewWildApricotService(cfg, logger)
	dbService := NewDBService(db, cfg, logger)
	StartBackgroundDatabaseUpdate(waService, dbService, logger)

	if err := router.RunTLS(":443", cfg.CertFile, cfg.KeyFile); err != nil {
		logger.Fatalf("Failed to start HTTPS server: %v", err)
	}
}