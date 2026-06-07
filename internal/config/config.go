package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/multitemplate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GetDSN returns the PostgreSQL connection string from env vars
func GetDSN() string {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbURL := os.Getenv("DB_URL")

	return fmt.Sprintf(
		"host=%s user=%s password=%s port=%s dbname=%s sslmode=require TimeZone=Asia/Jakarta",
		dbURL, dbUser, dbPassword, dbPort, dbName,
	)
}

// ConnectDB opens a GORM DB connection
func ConnectDB() (*gorm.DB, error) {
	return gorm.Open(postgres.Open(GetDSN()), &gorm.Config{})
}

// CORSConfig builds the CORS config from env vars
func CORSConfig() cors.Config {
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	corsOrigins := []string{"http://localhost:3000", "http://localhost:8080"}
	if allowedOrigins != "" {
		corsOrigins = strings.Split(allowedOrigins, ",")
	}

	cfg := cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// Allow all origins only when no ALLOWED_ORIGINS is set and AllowCredentials is false
	if os.Getenv("ALLOWED_ORIGINS") == "" && os.Getenv("ENV") == "development" {
		cfg.AllowAllOrigins = true
		cfg.AllowCredentials = false
	}

	return cfg
}

// SessionSecret returns the session key or a default
func SessionSecret() string {
	secret := os.Getenv("SESSION_SECRET_KEY")
	if secret == "" {
		secret = "CHANGE_ME_SESSION_KEY"
	}
	return secret
}

// Port returns the server port
func Port() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

// LoadTemplates loads multitemplate from the given directory
func LoadTemplates(templatesDir string) multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	layouts, err := filepath.Glob(templatesDir + "/layouts/*.html")
	if err != nil {
		return r
	}

	includes, err := filepath.Glob(templatesDir + "/**/*")
	if err != nil {
		return r
	}

	for _, include := range includes {
		layoutCopy := make([]string, len(layouts))
		copy(layoutCopy, layouts)
		files := append(layoutCopy, include)
		r.AddFromFiles(filepath.Base(include), files...)
	}
	return r
}