package models

import "gorm.io/gorm"

// User represents a user in the system
type User struct {
	gorm.Model
	Name     string `json:"name" binding:"required"`
	Email    string `gorm:"uniqueIndex" json:"email" binding:"required,email"`
	Password string `json:"-"` // Hashed password, excluded from JSON output
	Expenses []Expense // One-to-many relationship
}

// PublicUser is a DTO used for responses where the password must be omitted
type PublicUser struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
