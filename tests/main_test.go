// test/main_test.go
package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type CreateURLRequest struct {
	URL            string `json:"url"`
	CustomSlug     string `json:"customSlug,omitempty"`
	ExpirationDate string `json:"expirationDate,omitempty"`
}

type URLResponse struct {
	OriginalURL    string `json:"originalUrl"`
	ShortLink      string `json:"shortLink"`
	ExpirationDate string `json:"expirationDate"`
}

func TestURLShortener(t *testing.T) {
	// Test creating a short URL
	t.Run("Create Short URL", func(t *testing.T) {
		payload := CreateURLRequest{
			URL:            "https://www.google.com",
			ExpirationDate: "2024-12-31",
		}

		jsonData, err := json.Marshal(payload)
		assert.NoError(t, err)

		resp, err := http.Post(
			"http://localhost:8080/generate/shortlink",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response URLResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.ShortLink)
		assert.Equal(t, payload.URL, response.OriginalURL)

		// Store shortLink for subsequent tests
		shortLink := response.ShortLink

		// Test redirect
		redirectResp, err := http.Get("http://localhost:8080/" + shortLink)
		assert.NoError(t, err)
		defer redirectResp.Body.Close()
		assert.Equal(t, http.StatusFound, redirectResp.StatusCode)

		// Test deletion
		deleteReq, err := http.NewRequest(http.MethodDelete, "http://localhost:8080/"+shortLink, nil)
		assert.NoError(t, err)
		deleteResp, err := http.DefaultClient.Do(deleteReq)
		assert.NoError(t, err)
		defer deleteResp.Body.Close()
		assert.Equal(t, http.StatusOK, deleteResp.StatusCode)
	})

	// Test invalid URL
	t.Run("Invalid URL", func(t *testing.T) {
		payload := CreateURLRequest{
			URL: "invalid-url",
		}

		jsonData, err := json.Marshal(payload)
		assert.NoError(t, err)

		resp, err := http.Post(
			"http://localhost:8080/generate/shortlink",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Test custom slug
	t.Run("Custom Slug", func(t *testing.T) {
		payload := CreateURLRequest{
			URL:        "https://www.google.com",
			CustomSlug: "testslug",
		}

		jsonData, err := json.Marshal(payload)
		assert.NoError(t, err)

		resp, err := http.Post(
			"http://localhost:8080/generate/shortlink",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response URLResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, payload.CustomSlug, response.ShortLink)
	})
}
