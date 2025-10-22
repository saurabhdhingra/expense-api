package handlers

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"expense-api/config"
	"expense-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Helper function to convert GORM model to DTO
func ToExpenseDTO(e models.Expense) models.ExpenseDTO {
	return models.ExpenseDTO{
		ID: e.ID,
		Description: e.Description,
		Amount: e.Amount,
		Category: e.Category,
		Date: e.Date,
	}
}

// CreateExpense handles POST /expenses
func CreateExpense(c *gin.Context) {
	var expense models.Expense
	if err := c.ShouldBindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	// Validate Category
	if !models.ValidCategories[expense.Category] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category", "valid_categories": models.ValidCategories})
		return
	}

	// Set the UserID from the authenticated user
	expense.UserID = ExtractUserID(c)

	if err := config.DB.Create(&expense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create expense"})
		return
	}

	c.JSON(http.StatusCreated, ToExpenseDTO(expense))
}

// UpdateExpense handles PUT /expenses/:id
func UpdateExpense(c *gin.Context) {
	expenseIDStr := c.Param("id")
	expenseID, err := strconv.ParseUint(expenseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Expense ID"})
		return
	}

	userID := ExtractUserID(c)

	var expense models.Expense
	// Check ownership and existence
	if err := config.DB.Where("id = ? AND user_id = ?", expenseID, userID).First(&expense).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden or Not Found"}) 
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Category Validation on update
	if category, ok := updateData["category"].(string); ok {
		if !models.ValidCategories[category] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category in update", "valid_categories": models.ValidCategories})
			return
		}
	}

	// GORM Updates handles partial updates and data type safety
	if err := config.DB.Model(&expense).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update expense"})
		return
	}

	config.DB.First(&expense, expense.ID)
	c.JSON(http.StatusOK, ToExpenseDTO(expense))
}

// DeleteExpense handles DELETE /expenses/:id
func DeleteExpense(c *gin.Context) {
	expenseIDStr := c.Param("id")
	expenseID, err := strconv.ParseUint(expenseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Expense ID"})
		return
	}
	
	userID := ExtractUserID(c)
	
	var expense models.Expense
	// Check ownership and existence
	if err := config.DB.Where("id = ? AND user_id = ?", expenseID, userID).First(&expense).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden or Not Found"}) 
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if err := config.DB.Delete(&expense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete expense"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListExpenses handles GET /expenses?filter=past_week
func ListExpenses(c *gin.Context) {
	userID := ExtractUserID(c)
	filter := c.DefaultQuery("filter", "all")

	// Custom date range parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	// Determine date range based on filter type
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	switch filter {
	case "past_week":
		startDate = today.AddDate(0, 0, -7)
		endDate = today.AddDate(0, 0, 1) // To include all of today
	case "past_month":
		startDate = today.AddDate(0, -1, 0)
		endDate = today.AddDate(0, 0, 1)
	case "last_3_months":
		startDate = today.AddDate(0, -3, 0)
		endDate = today.AddDate(0, 0, 1)
	case "custom":
		if startDateStr == "" || endDateStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Custom filter requires start_date and end_date in YYYY-MM-DD format"})
			return
		}
		// Parse dates: assuming YYYY-MM-DD format
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
			return
		}
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
			return
		}
		// Add one day to end date to include all transactions on the end day
		endDate = endDate.AddDate(0, 0, 1)
	case "all":
		// No date filter applied
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter type"})
		return
	}

	var expenses []models.Expense
	query := config.DB.Where("user_id = ?", userID).Order("date desc")

	// Apply date range filter if not "all"
	if filter != "all" {
		query = query.Where("date >= ? AND date < ?", startDate, endDate)
	}

	if err := query.Find(&expenses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch expenses"})
		return
	}

	var responseData []models.ExpenseDTO
	for _, exp := range expenses {
		responseData = append(responseData, ToExpenseDTO(exp))
	}

	c.JSON(http.StatusOK, responseData)
}

// GetAnalytics handles GET /analytics
func GetAnalytics(c *gin.Context) {
	userID := ExtractUserID(c)
	now := time.Now().UTC()
	
	// 1. Calculate time boundaries for the current period
	// Start of Day
	// startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	// Start of Month
	// startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	// Start of Year
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)


	// 2. Fetch all expenses for the current year
	var expensesInYear []models.Expense
	if err := config.DB.Where("user_id = ? AND date >= ?", userID, startOfYear).Order("date desc").Find(&expensesInYear).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data for analytics"})
		return
	}

	// 3. Process Data for Analytics
	analytics := models.AnalyticsResponse{
		Daily:   make([]models.AnalyticsItem, 0),
		Monthly: make([]models.AnalyticsItem, 0),
		Yearly:  make([]models.AnalyticsItem, 0),
	}

	// Grouping maps
	dailyMap := make(map[string][]models.Expense)   // Key: YYYY-MM-DD
	monthlyMap := make(map[string][]models.Expense) // Key: YYYY-MM
	yearlyMap := make(map[string][]models.Expense)  // Key: YYYY

	for _, exp := range expensesInYear {
		// Daily grouping
		dailyKey := exp.Date.Format("2006-01-02")
		dailyMap[dailyKey] = append(dailyMap[dailyKey], exp)

		// Monthly grouping
		monthlyKey := exp.Date.Format("2006-01")
		monthlyMap[monthlyKey] = append(monthlyMap[monthlyKey], exp)

		// Yearly grouping
		yearlyKey := exp.Date.Format("2006")
		yearlyMap[yearlyKey] = append(yearlyMap[yearlyKey], exp)
	}

	// Helper to calculate total and top 5
	processGroup := func(group map[string][]models.Expense, periodFormat string) []models.AnalyticsItem {
		var items []models.AnalyticsItem
		for key, expList := range group {
			total := 0.0
			var dtos []models.ExpenseDTO
			
			// 1. Calculate Total
			for _, exp := range expList {
				total += exp.Amount
				dtos = append(dtos, ToExpenseDTO(exp))
			}

			// 2. Sort DTOs by Amount Descending (Top 5)
			sort.Slice(dtos, func(i, j int) bool {
				return dtos[i].Amount > dtos[j].Amount
			})
			
			topN := 5
			if len(dtos) < topN {
				topN = len(dtos)
			}

			// Format period key (e.g., convert "2023-10" to "October 2023")
			var periodDisplay string
			if periodFormat == "2006-01-02" {
				periodDisplay = key // YYYY-MM-DD
			} else if periodFormat == "2006-01" {
				t, _ := time.Parse("2006-01", key)
				periodDisplay = t.Format("January 2006")
			} else {
				periodDisplay = key // YYYY
			}

			items = append(items, models.AnalyticsItem{
				Period: periodDisplay,
				TotalExpenses: total,
				TopTransactions: dtos[:topN],
			})
		}
		// Sort by period key for chronological order
		sort.Slice(items, func(i, j int) bool {
			return items[i].Period < items[j].Period
		})
		return items
	}

	// Populate analytics response
	analytics.Daily = processGroup(dailyMap, "2006-01-02")
	analytics.Monthly = processGroup(monthlyMap, "2006-01")
	analytics.Yearly = processGroup(yearlyMap, "2006")

	c.JSON(http.StatusOK, analytics)
}