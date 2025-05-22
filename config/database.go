package config

import (
	"fmt"
	"os"
	"url-shortener/logging"
	"url-shortener/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// SetupDatabase initializes and returns a database connection
func SetupDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logging.Log.WithError(err).Error("Failed to connect to database")
		return nil, err
	}

	// Use models.URL for AutoMigrate
	err = db.AutoMigrate(&models.URL{})
	if err != nil {
		logging.Log.WithError(err).Error("Failed to migrate database")
		return nil, err
	}
	logging.Log.Info("Database migration completed successfully")
	return db, nil
}
