package models

import (
	"time"

	"gorm.io/gorm"
)

type URL struct {
	gorm.Model
	OriginalURL    string    `gorm:"type:text;not null"`
	ShortLink      string    `gorm:"type:varchar(10);unique;not null"`
	ExpirationDate time.Time `gorm:"not null"`
}
