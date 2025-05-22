package main

import (
	"fmt"
	"os"
	"time"
	"url-shortener/config" // Added for SetupDatabase
	"url-shortener/controllers"
	"url-shortener/logging" // Added for logrus
	"url-shortener/middleware"
	"url-shortener/models" // models is used by config.SetupDatabase if it still does AutoMigrate

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	// "gorm.io/driver/mysql" // No longer needed directly in main.go
	"gorm.io/gorm"
)

// Struct definitions (URL, CreateURLRequest, URLResponse) are removed.
// generateRandomSlug function is removed (assuming it's in utils/random.go).
// setupDatabase function is removed (moved to config/database.go).

func setupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// Apply RequestLogger middleware globally - should be one of the first
	router.Use(middleware.RequestLogger())
	// Apply RateLimiter middleware globally
	router.Use(middleware.RateLimiter())
	// Apply SecurityHeaders middleware globally
	router.Use(middleware.SecurityHeaders())

	// Initialize controller
	urlController := controllers.NewURLController(db)

	// Add ping endpoint for health check
	router.GET("/ping", urlController.Ping) // Use the Ping method from URLController

	// Setup routes
	router.POST("/generate/shortlink", urlController.CreateShortURL)
	router.GET("/:shortLink", urlController.RedirectToURL)
	router.DELETE("/:shortLink", urlController.DeleteShortURL)
	// Removed direct handler implementations

	return router
}

func main() {
	// Initialize random seed (if generateRandomSlug was used directly in main and not in utils)
	// rand.Seed(time.Now().UnixNano()) // This is now handled in utils/random.go's init

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logging.Log.WithError(err).Fatal("Error loading .env file")
	}

	// Setup database
	db, err := config.SetupDatabase() // Use config.SetupDatabase
	if err != nil {
		logging.Log.WithError(err).Fatal("Database setup failed")
	}

	// Setup router
	router := setupRouter(db)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	logging.Log.Infof("Starting server on port %s", port)
	// Start server
	if err := router.Run(":" + port); err != nil {
		logging.Log.WithError(err).Fatal("Failed to start server")
	}
}
