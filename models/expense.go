package models

import (
	"time"

	"gorm.io/gorm"
)

var ValidCategories = map[string]bool{
	"Groceries": true,
	"Leisure": true,
	"Electronics": true,
	"Utilities": true,
	"Clothing": true,
	"Health": true,
	"Others": true,
}

type Expense struct {
	gorm.Model
	Description 	string		`json:"description" binding:"required"`
	Amount			float64		`json:"amount" binding:"required,gt=0"`
	Category		string		`json:"category" binding:"required"`
	Date 			time.Time	`json:"date" binding:"required"`
	UserID			uint		`json:"user_id"`
	User			User
}

type ExpenseDTO struct {
	ID 				uint		`json:"id"`
	Description		string		`json:"description"`
	Amount			float64		`json:"amount"`
	Category		string		`json:"category"`
	Date			time.Time	`json:"date"`
}

type AnalyticsItem struct {
	Period			string			`json:"period"`
	TotalExpenses	float64			`json:"total_expenses"`
	TopTransactions	[]ExpenseDTO	`json:"top_5"`
}

type AnalyticsResponse struct {
	Daily		[]AnalyticsItem		`json:"daily"`
	Monthly		[]AnalyticsItem		`json:"monthly"`
	Yearly		[]AnalyticsItem		`json:"yearly"`
}