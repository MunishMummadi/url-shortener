package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type URL struct {
	gorm.Model
	OriginalURL    string    `gorm:"type:text;not null"`
	ShortLink      string    `gorm:"type:varchar(10);unique;not null"`
	ExpirationDate time.Time `gorm:"not null"`
}

type CreateURLRequest struct {
	URL            string `json:"url" binding:"required"`
	CustomSlug     string `json:"customSlug"`
	ExpirationDate string `json:"expirationDate"`
}

type URLResponse struct {
	OriginalURL    string    `json:"originalUrl"`
	ShortLink      string    `json:"shortLink"`
	ExpirationDate time.Time `json:"expirationDate"`
}

func generateRandomSlug(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db.AutoMigrate(&URL{})

	router := gin.Default()

	// Generate short URL
	router.POST("/generate/shortlink", func(c *gin.Context) {
		var request CreateURLRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// Validate URL
		if !strings.HasPrefix(request.URL, "http://") && !strings.HasPrefix(request.URL, "https://") {
			c.JSON(400, gin.H{"error": "URL must start with http:// or https://"})
			return
		}

		// Set expiration date
		expirationDate := time.Now().Add(24 * time.Hour) // Default 24 hours
		if request.ExpirationDate != "" {
			parsedDate, err := time.Parse("2006-01-02", request.ExpirationDate)
			if err == nil {
				expirationDate = parsedDate
			}
		}

		var shortLink string
		if request.CustomSlug != "" {
			// Check if custom slug exists
			var existingURL URL
			if result := db.Where("short_link = ?", request.CustomSlug).First(&existingURL); result.Error == nil {
				c.JSON(409, gin.H{"error": "Custom slug already exists"})
				return
			}
			shortLink = request.CustomSlug
		} else {
			// Generate random slug
			shortLink = generateRandomSlug(6)
			for {
				var existingURL URL
				if result := db.Where("short_link = ?", shortLink).First(&existingURL); result.Error != nil {
					break
				}
				shortLink = generateRandomSlug(6)
			}
		}

		url := URL{
			OriginalURL:    request.URL,
			ShortLink:      shortLink,
			ExpirationDate: expirationDate,
		}

		result := db.Create(&url)
		if result.Error != nil {
			c.JSON(500, gin.H{"error": "Failed to create short URL"})
			return
		}

		c.JSON(201, URLResponse{
			OriginalURL:    url.OriginalURL,
			ShortLink:      url.ShortLink,
			ExpirationDate: url.ExpirationDate,
		})
	})

	// Redirect to original URL
	router.GET("/:shortLink", func(c *gin.Context) {
		shortLink := c.Param("shortLink")
		var url URL

		if result := db.Where("short_link = ?", shortLink).First(&url); result.Error != nil {
			c.JSON(404, gin.H{"error": "Short URL not found"})
			return
		}

		if time.Now().After(url.ExpirationDate) {
			db.Delete(&url)
			c.JSON(410, gin.H{"error": "URL has expired"})
			return
		}

		c.Redirect(302, url.OriginalURL)
	})

	// Delete URL
	router.DELETE("/:shortLink", func(c *gin.Context) {
		shortLink := c.Param("shortLink")
		var url URL

		if result := db.Where("short_link = ?", shortLink).First(&url); result.Error != nil {
			c.JSON(404, gin.H{"error": "Short URL not found"})
			return
		}

		db.Delete(&url)
		c.JSON(200, gin.H{"message": "URL deleted successfully"})
	})

	router.Run(":8080")
}
