// controllers/url_controller.go
package controllers

import (
	"net/http"
	"time"
	"url-shortener/dto/request"
	"url-shortener/dto/response"
	"url-shortener/services"

	"github.com/gin-gonic/gin"
)

type URLController struct {
	urlService services.URLService
}

func NewURLController(urlService services.URLService) *URLController {
	return &URLController{urlService: urlService}
}

func (c *URLController) GenerateShortLink(ctx *gin.Context) {
	var req request.CreateURLRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse{
			Status: "400",
			Error:  err.Error(),
		})
		return
	}

	// Parse expiration date or set default (24 hours from now)
	expirationDate := time.Now().Add(24 * time.Hour)
	if req.ExpirationDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.ExpirationDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, response.ErrorResponse{
				Status: "400",
				Error:  "Invalid expiration date format",
			})
			return
		}
		expirationDate = parsedDate
	}

	url, err := c.urlService.CreateURL(req.URL, req.CustomSlug, expirationDate)
	if err != nil {
		ctx.JSON(http.StatusConflict, response.ErrorResponse{
			Status: "409",
			Error:  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, response.URLResponse{
		OriginalURL:    url.OriginalURL,
		ShortLink:      url.ShortLink,
		ExpirationDate: url.ExpirationDate,
	})
}

func (c *URLController) ValidateCustomSlug(ctx *gin.Context) {
	var req request.ValidateSlugRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse{
			Status: "400",
			Error:  err.Error(),
		})
		return
	}

	exists := c.urlService.IsCustomSlugExists(req.CustomSlug)
	if exists {
		ctx.JSON(http.StatusConflict, response.MessageResponse{
			Message: "Custom slug already exists",
		})
		return
	}

	ctx.JSON(http.StatusOK, response.MessageResponse{
		Message: "Custom slug is available",
	})
}

func (c *URLController) RedirectToOriginalURL(ctx *gin.Context) {
	shortLink := ctx.Param("shortLink")
	url, err := c.urlService.GetURL(shortLink)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.ErrorResponse{
			Status: "404",
			Error:  "URL does not exist or it might have expired",
		})
		return
	}

	ctx.Redirect(http.StatusFound, url.OriginalURL)
}

func (c *URLController) DeleteShortURL(ctx *gin.Context) {
	shortLink := ctx.Param("shortLink")
	err := c.urlService.DeleteURL(shortLink)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.ErrorResponse{
			Status: "404",
			Error:  "URL does not exist or it might have expired",
		})
		return
	}

	ctx.JSON(http.StatusOK, response.MessageResponse{
		Message: "Short URL has been successfully deleted",
	})
}

func (c *URLController) SetExpirationDate(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")
	var req request.UpdateExpirationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse{
			Status: "400",
			Error:  err.Error(),
		})
		return
	}

	expirationDate, _ := time.Parse("2006-01-02", req.ExpirationDate)
	url, err := c.urlService.SetExpirationDate(shortCode, expirationDate)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.ErrorResponse{
			Status: "404",
			Error:  "URL does not exist or it might have expired",
		})
		return
	}

	ctx.JSON(http.StatusOK, response.URLResponse{
		OriginalURL:    url.OriginalURL,
		ShortLink:      url.ShortLink,
		ExpirationDate: url.ExpirationDate,
	})
}
