// main_test.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var testDB *gorm.DB
var testRouter *gin.Engine

func TestMain(m *testing.M) {
	// Set test environment
	gin.SetMode(gin.TestMode)

	// Setup test database
	setupTestDB()

	// Run tests
	code := m.Run()

	// Cleanup
	cleanupTestDB()

	os.Exit(code)
}

func setupTestDB() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	testDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	testDB.AutoMigrate(&URL{})
	testRouter = setupRouter(testDB) // Pass testDB to setupRouter
}

func cleanupTestDB() {
	// Clean up test data
	sqlDB, err := testDB.DB()
	if err != nil {
		panic("failed to get underlying sql.DB")
	}
	sqlDB.Exec("DELETE FROM urls")
}

func TestCreateShortURL(t *testing.T) {
	cleanupTestDB() // Clean before each test

	t.Run("Valid URL", func(t *testing.T) {
		payload := CreateURLRequest{
			URL:            "https://www.google.com",
			ExpirationDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		}

		jsonData, err := json.Marshal(payload)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/generate/shortlink", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		testRouter.ServeHTTP(w, req)

		assert.Equal(t, 201, w.Code)

		var response URLResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.ShortLink)
	})

	t.Run("Invalid URL", func(t *testing.T) {
		payload := CreateURLRequest{
			URL: "invalid-url",
		}

		jsonData, err := json.Marshal(payload)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/generate/shortlink", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		testRouter.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
	})
}

func TestRedirectURL(t *testing.T) {
	cleanupTestDB() // Clean before test

	// First create a URL
	payload := CreateURLRequest{
		URL:        "https://www.google.com",
		CustomSlug: "testredirect",
	}

	jsonData, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/generate/shortlink", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	// Test redirection
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/testredirect", nil)
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, 302, w.Code)
	assert.Equal(t, "https://www.google.com", w.Header().Get("Location"))
}

func TestDeleteURL(t *testing.T) {
	cleanupTestDB() // Clean before test

	// First create a URL
	payload := CreateURLRequest{
		URL:        "https://www.google.com",
		CustomSlug: "testdelete",
	}

	jsonData, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/generate/shortlink", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	// Delete the URL
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/testdelete", nil)
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// Verify deletion
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/testdelete", nil)
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}
