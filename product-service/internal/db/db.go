package db

import (
	"product-service/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dsn := config.Config.DatabaseURL
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,                                // turn off transaction for all of query, increase 30% performance
		Logger:                 logger.Default.LogMode(logger.Info), // Log all SQL queries
	})
	if err != nil {
		panic("Failed to connect to database")
	}
	DB = db

	// Auto-migrate models
	db.AutoMigrate()
}
