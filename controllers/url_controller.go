package controllers

import (
	"net/http"
	"strings"
	"time"
	"url-shortener/dto/request"
	"url-shortener/dto/response"
	"url-shortener/logging" // Added for logrus
	"url-shortener/models"
	"url-shortener/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus" // Added for logrus fields
	"gorm.io/gorm"
)

// Helper function for standardized error responses
func errorResponse(c *gin.Context, statusCode int, message string) {
	logging.Log.WithFields(logrus.Fields{
		"status_code": statusCode,
		"path":        c.Request.URL.Path,
		"method":      c.Request.Method,
	}).Error(message)
	c.JSON(statusCode, gin.H{"error": message})
}

// Helper function for internal server errors (logs actual error, returns generic message)
func internalServerErrorResponse(c *gin.Context, err error, message string) {
	logging.Log.WithError(err).WithFields(logrus.Fields{
		"path":   c.Request.URL.Path,
		"method": c.Request.Method,
	}).Error(message) // Log the detailed error internally
	c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"}) // Generic message to client
}

type URLController struct {
	db *gorm.DB
}

func NewURLController(db *gorm.DB) *URLController {
	return &URLController{db: db}
}

func (controller *URLController) CreateShortURL(c *gin.Context) {
	var req request.CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		errorResponse(c, http.StatusBadRequest, "URL must start with http:// or https://")
		return
	}

	expirationDate := time.Now().Add(24 * time.Hour) // Default expiration: 24 hours
	if req.ExpirationDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.ExpirationDate)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid expiration date format. Use YYYY-MM-DD")
			return
		}
		expirationDate = parsedDate
	}

	var shortLink string
	if req.CustomSlug != "" {
		var existingURL models.URL
		// Check if custom slug already exists
		if result := controller.db.Where("short_link = ?", req.CustomSlug).First(&existingURL); result.Error == nil {
			errorResponse(c, http.StatusConflict, "Custom slug already exists")
			return
		} else if result.Error != gorm.ErrRecordNotFound {
			internalServerErrorResponse(c, result.Error, "Database error checking custom slug")
			return
		}
		shortLink = req.CustomSlug
	} else {
		// Generate a unique random slug
		for {
			shortLink = utils.GenerateRandomSlug(6)
			var existingURL models.URL
			if result := controller.db.Where("short_link = ?", shortLink).First(&existingURL); result.Error != nil {
				if result.Error == gorm.ErrRecordNotFound {
					// If error (record not found), slug is unique
					break
				}
				internalServerErrorResponse(c, result.Error, "Database error generating unique slug")
				return
			}
		}
	}

	url := models.URL{
		OriginalURL:    req.URL,
		ShortLink:      shortLink,
		ExpirationDate: expirationDate,
	}

	result := controller.db.Create(&url)
	if result.Error != nil {
		internalServerErrorResponse(c, result.Error, "Failed to create short URL")
		return
	}

	c.JSON(http.StatusCreated, response.URLResponse{
		OriginalURL:    url.OriginalURL,
		ShortLink:      url.ShortLink,
		ExpirationDate: url.ExpirationDate,
	})
}

func (controller *URLController) RedirectToURL(c *gin.Context) {
	shortLink := c.Param("shortLink")
	var url models.URL

	if result := controller.db.Where("short_link = ?", shortLink).First(&url); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			errorResponse(c, http.StatusNotFound, "Short URL not found")
		} else {
			internalServerErrorResponse(c, result.Error, "Database error fetching URL for redirect")
		}
		return
	}

	// Check if the URL has expired
	if time.Now().After(url.ExpirationDate) {
		if delResult := controller.db.Delete(&url); delResult.Error != nil {
			internalServerErrorResponse(c, delResult.Error, "Failed to delete expired URL")
			// Even if deletion fails, the URL is still expired from client's perspective
			// So, we still return StatusGone.
			// Alternatively, one might choose to not delete and let a background job handle it.
		}
		errorResponse(c, http.StatusGone, "URL has expired")
		return
	}

	c.Redirect(http.StatusFound, url.OriginalURL)
}

func (controller *URLController) DeleteShortURL(c *gin.Context) {
	shortLink := c.Param("shortLink")
	var url models.URL

	// Find the URL by shortLink
	if result := controller.db.Where("short_link = ?", shortLink).First(&url); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			errorResponse(c, http.StatusNotFound, "Short URL not found")
		} else {
			internalServerErrorResponse(c, result.Error, "Database error fetching URL for deletion")
		}
		return
	}

	// Delete the URL
	if result := controller.db.Delete(&url); result.Error != nil {
		internalServerErrorResponse(c, result.Error, "Failed to delete URL")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "URL deleted successfully"})
}

func (controller *URLController) Ping(c *gin.Context) {
	sqlDB, err := controller.db.DB() // controller is the URLController instance
	if err != nil {
		// Log the error internally
		logging.Log.WithError(err).Error("Failed to get underlying DB object for health check")
		// Use the standardized error response
		internalServerErrorResponse(c, err, "Error accessing database for health check") // Use internalServerErrorResponse for 500
		return
	}

	err = sqlDB.Ping()
	if err != nil {
		// Log the error internally
		logging.Log.WithError(err).Warn("Database ping failed during health check")
		// Use the standardized error response, but with 503
		errorResponse(c, http.StatusServiceUnavailable, "Database not reachable")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"status":  "Database connected successfully", // This message remains the same on success
	})
}
