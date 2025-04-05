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
	// Init Default  Roles
	initRole()
	// Init Default  Employee
	initEmployee()
}
func initRole() {
	var count int64
	DB.Model(models.Role{}).Count(&count)
	if count > 0 {
		fmt.Println("Data is exit.")
		return
	}
	sql := "INSERT INTO `roles` VALUES ('6ac14e4c-4239-4f0d-8fcf-c0441173e971','2025-04-05 12:26:46.358','2025-04-05 12:26:46.358','manager','manager'),('b728e6b1-8925-4a35-aa81-69d2e83a0c04','2025-04-05 12:26:46.363','2025-04-05 12:26:46.363','employee','employee'),('ccaa33ad-7268-4989-aa6c-61f04e741475','2025-04-05 12:26:46.328','2025-04-05 12:26:46.328','admin','admin');"
	DB.Exec(sql)
}
func initEmployee() {
	var count int64
	DB.Model(models.Employee{}).Count(&count)
	if count > 0 {
		fmt.Println("Data is exit.")
		return
	}
	sql := "INSERT INTO `employees` VALUES ('ee1b2de9-c8b6-436b-8527-011a036f5fbb','2025-04-05 12:26:46.446','2025-04-05 12:26:46.446','admin','$2a$10$wN0iakb2V4hRfXahk0UgWejh7JRysRjePxTjyAOnlxbRrkIsgUK5C','000000000','Trần Thanh Hoàng','Hoangila2016@gmail.com','2003-09-16 07:00:00.000','Backend','090999999','https://facebook.com.vn',1,'ccaa33ad-7268-4989-aa6c-61f04e741475');"
	DB.Exec(sql)
}
