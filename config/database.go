// config/database.go
package config

import (
	"fmt"
	"log"
	"os"
	"url-shortener/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func SetupDatabase() *gorm.DB {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&models.URL{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	return db
}
