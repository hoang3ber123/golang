package db

import (
	"auth-service/config"
	"auth-service/internal/models"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.Config.DatabaseUsername, config.Config.DatabasePassword, config.Config.DatabaseHost, config.Config.DatabasePort, config.Config.DatabaseName)
	fmt.Println(dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,                                // turn off transaction for all of query, increase 30% performance
		Logger:                 logger.Default.LogMode(logger.Info), // Log all SQL queries
	})
	if err != nil {
		panic("Failed to connect to database")
	}
	DB = db

	// Auto-migrate models
	db.AutoMigrate(&models.User{}, &models.Role{}, &models.Employee{})
}
