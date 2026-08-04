package database

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	SERVER_PORT string
	DB_PASS     string
	DB_USER     string
	DB_NAME     string
	DB_HOST     string
	DB_PORT     string
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		SERVER_PORT: os.Getenv("SERVER_PORT"),
		DB_PASS:     os.Getenv("DB_PASS"),
		DB_USER:     os.Getenv("DB_USER"),
		DB_NAME:     os.Getenv("DB_NAME"),
		DB_HOST:     os.Getenv("DB_HOST"),
		DB_PORT:     os.Getenv("DB_PORT"),
	}
}

var DB *gorm.DB

func Connect(config *DatabaseConfig) (err error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", config.DB_HOST, config.DB_USER, config.DB_PASS, config.DB_NAME, config.DB_PORT)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	})
	if err != nil {
		return err
	}

	return nil
}
