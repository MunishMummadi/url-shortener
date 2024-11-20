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

func setupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// Add ping endpoint for health check
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "Database connected successfully",
		})
	})

	// Generate short URL
	router.POST("/generate/shortlink", func(c *gin.Context) {
		var request CreateURLRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if !strings.HasPrefix(request.URL, "http://") && !strings.HasPrefix(request.URL, "https://") {
			c.JSON(400, gin.H{"error": "URL must start with http:// or https://"})
			return
		}

		expirationDate := time.Now().Add(24 * time.Hour)
		if request.ExpirationDate != "" {
			parsedDate, err := time.Parse("2006-01-02", request.ExpirationDate)
			if err == nil {
				expirationDate = parsedDate
			}
		}

		var shortLink string
		if request.CustomSlug != "" {
			var existingURL URL
			if result := db.Where("short_link = ?", request.CustomSlug).First(&existingURL); result.Error == nil {
				c.JSON(409, gin.H{"error": "Custom slug already exists"})
				return
			}
			shortLink = request.CustomSlug
		} else {
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

	return router
}

func setupDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&URL{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Setup database
	db, err := setupDatabase()
	if err != nil {
		log.Fatal("Database setup failed:", err)
	}

	// Setup router
	router := setupRouter(db)

	// Start server
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
