package repositories

import (
	"url-shortener/models"

	"gorm.io/gorm"
)

type URLRepository interface {
	Create(url *models.URL) error
	FindByShortLink(shortLink string) (*models.URL, error)
	Delete(url *models.URL) error
	ExistsByShortLink(shortLink string) bool
	Update(url *models.URL) error
}

type urlRepository struct {
	db *gorm.DB
}

func NewURLRepository(db *gorm.DB) URLRepository {
	return &urlRepository{db: db}
}

func (r *urlRepository) Create(url *models.URL) error {
	return r.db.Create(url).Error
}

func (r *urlRepository) FindByShortLink(shortLink string) (*models.URL, error) {
	var url models.URL
	if err := r.db.Where("short_link = ?", shortLink).First(&url).Error; err != nil {
		return nil, err
	}
	return &url, nil
}

func (r *urlRepository) Delete(url *models.URL) error {
	return r.db.Delete(url).Error
}

func (r *urlRepository) ExistsByShortLink(shortLink string) bool {
	var count int64
	r.db.Model(&models.URL{}).Where("short_link = ?", shortLink).Count(&count)
	return count > 0
}

func (r *urlRepository) Update(url *models.URL) error {
	return r.db.Save(url).Error
}
