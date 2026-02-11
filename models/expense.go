package models

import (
	"time"

	"gorm.io/gorm"
)

var ValidCategories = map[string]bool{
	"Groceries":   true,
	"Leisure":     true,
	"Electronics": true,
	"Utilities":   true,
	"Clothing":    true,
	"Health":      true,
	"Others":      true,
}

type Expense struct {
	gorm.Model
	Description string    `json:"description" binding:"required"`
	Amount      float64   `json:"amount" binding:"required,gt=0"`
	Category    string    `json:"category" binding:"required"`
	Date        time.Time `json:"date" binding:"required"`
	UserID      uint      `json:"user_id"`
	User        User      `json:"-"`
	WalletID    uint      `json:"wallet_id"`
	Wallet      Wallet    `json:"-"`
}

type ExpenseDTO struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Date        time.Time `json:"date"`
	WalletID    string    `json:"wallet_id"`
}

type AnalyticsItem struct {
	Period          string       `json:"period"`
	TotalExpenses   float64      `json:"total_expenses"`
	TopTransactions []ExpenseDTO `json:"top_5"`
}

type AnalyticsResponse struct {
	Daily   []AnalyticsItem `json:"daily"`
	Monthly []AnalyticsItem `json:"monthly"`
	Yearly  []AnalyticsItem `json:"yearly"`
}

type CategoryDistribution struct {
	Category    string  `json:"category"`
	TotalAmount float64 `json:"totalAmount"`
}

type MonthlyExpenseGroup struct {
	Month    string       `json:"month"`
	Expenses []ExpenseDTO `json:"expenses"`
}

type ExpenseResponseData struct {
	MonthlyExpenses          []MonthlyExpenseGroup  `json:"monthlyExpenses"`
	CurrentMonthDistribution []CategoryDistribution `json:"currentMonthDistribution"`
}

type BarChartDataPoint struct {
	ID     string  `json:"id"`
	Label  string  `json:"label"`
	Amount float64 `json:"amount"`
}

type InsightsResponse struct {
	ChartData       []BarChartDataPoint `json:"chartData"`
	TopTransactions []ExpenseDTO        `json:"topTransactions"`
}
