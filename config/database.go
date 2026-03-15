package config

import (
	"expense-api/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := "postgresql://neondb_owner:npg_MephZjqu91tc@ep-withered-mountain-a1nhzch9-pooler.ap-southeast-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Running database migrations...")
	err = db.AutoMigrate(&models.User{}, &models.Expense{}, &models.Wallet{})
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database connection successful and migrations complete.")

	DB = db
}
