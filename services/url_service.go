// services/url_service.go
package services

import (
	"errors"
	"time"
	"url-shortener/models"
	"url-shortener/repositories"
	"url-shortener/utils"
)

type URLService interface {
	CreateURL(originalURL, customSlug string, expirationDate time.Time) (*models.URL, error)
	GetURL(shortLink string) (*models.URL, error)
	DeleteURL(shortLink string) error
	IsCustomSlugExists(customSlug string) bool
	SetExpirationDate(shortLink string, expirationDate time.Time) (*models.URL, error)
}

type urlService struct {
	urlRepo repositories.URLRepository
}

func NewURLService(urlRepo repositories.URLRepository) URLService {
	return &urlService{urlRepo: urlRepo}
}

func (s *urlService) CreateURL(originalURL, customSlug string, expirationDate time.Time) (*models.URL, error) {
	var shortLink string
	if customSlug != "" {
		if s.urlRepo.ExistsByShortLink(customSlug) {
			return nil, errors.New("custom slug already exists")
		}
		shortLink = customSlug
	} else {
		shortLink = utils.GenerateRandomSlug(6)
		// Keep generating until we find a unique slug
		for s.urlRepo.ExistsByShortLink(shortLink) {
			shortLink = utils.GenerateRandomSlug(6)
		}
	}

	url := &models.URL{
		OriginalURL:    originalURL,
		ShortLink:      shortLink,
		CreationDate:   time.Now(),
		ExpirationDate: expirationDate,
	}

	if err := s.urlRepo.Create(url); err != nil {
		return nil, err
	}

	return url, nil
}

func (s *urlService) GetURL(shortLink string) (*models.URL, error) {
	url, err := s.urlRepo.FindByShortLink(shortLink)
	if err != nil {
		return nil, err
	}

	if url.ExpirationDate.Before(time.Now()) {
		s.urlRepo.Delete(url)
		return nil, errors.New("url has expired")
	}

	return url, nil
}

func (s *urlService) DeleteURL(shortLink string) error {
	url, err := s.urlRepo.FindByShortLink(shortLink)
	if err != nil {
		return err
	}
	return s.urlRepo.Delete(url)
}

func (s *urlService) IsCustomSlugExists(customSlug string) bool {
	return s.urlRepo.ExistsByShortLink(customSlug)
}

func (s *urlService) SetExpirationDate(shortLink string, expirationDate time.Time) (*models.URL, error) {
	url, err := s.urlRepo.FindByShortLink(shortLink)
	if err != nil {
		return nil, err
	}

	url.ExpirationDate = expirationDate
	if err := s.urlRepo.Update(url); err != nil {
		return nil, err
	}

	return url, nil
}
