package handlers

import (
	"errors"
	"fmt"
	"log"
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
		ID:          fmt.Sprintf("%d", e.ID),
		Description: e.Description,
		Amount:      e.Amount,
		Category:    e.Category,
		Date:        e.Date,
		WalletID:    fmt.Sprintf("%d", e.WalletID),
	}
}

// CreateExpense handles POST /expenses
func CreateExpense(c *gin.Context) {
	var req models.ExpenseCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateExpense Binding Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	log.Printf("CreateExpense Request: Description=%s, UseAmount=%f, Category=%s, WalletID=%s", req.Description, req.Amount, req.Category, req.WalletID)

	// Validate Category
	if !models.ValidCategories[req.Category] {
		log.Printf("CreateExpense Error: Invalid category %s", req.Category)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category", "valid_categories": models.ValidCategories})
		return
	}

	// Set the UserID from the authenticated user
	userID := ExtractUserID(c)

	expense := models.Expense{
		Description: req.Description,
		Amount:      req.Amount,
		Category:    req.Category,
		Date:        req.Date,
		UserID:      userID,
	}

	// Validate WalletID and ownership
	if req.WalletID != "" {
		wID, err := strconv.ParseUint(req.WalletID, 10, 32)
		if err != nil {
			log.Printf("CreateExpense Error: Invalid WalletID format %s", req.WalletID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet_id format"})
			return
		}
		expense.WalletID = uint(wID)

		var wallet models.Wallet
		if err := config.DB.Where("id = ? AND user_id = ?", expense.WalletID, userID).First(&wallet).Error; err != nil {
			log.Printf("CreateExpense Error: Wallet %d not found or not owned by user %d", expense.WalletID, userID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet_id or you do not own this wallet"})
			return
		}
		// Update balance
		wallet.Balance -= expense.Amount
		config.DB.Save(&wallet)
	}

	if err := config.DB.Create(&expense).Error; err != nil {
		log.Printf("CreateExpense Error: Database failure: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create expense"})
		return
	}

	log.Printf("CreateExpense Success: ID=%d", expense.ID)
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

// ListExpenses handles GET /expenses?filter=past_week&wallet_id=1
func ListExpenses(c *gin.Context) {
	userID := ExtractUserID(c)
	filter := c.DefaultQuery("filter", "all")
	walletIDStr := c.Query("wallet_id")

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

	// Apply wallet_id filter if provided
	if walletIDStr != "" {
		wID, err := strconv.ParseUint(walletIDStr, 10, 32)
		if err == nil {
			query = query.Where("wallet_id = ?", uint(wID))
		}
	}

	// Apply date range filter if not "all"
	if filter != "all" {
		query = query.Where("date >= ? AND date < ?", startDate, endDate)
	}

	if err := query.Find(&expenses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch expenses"})
		return
	}

	// If wallet_id is provided, return flat array instead of grouped format
	if walletIDStr != "" {
		responseData := make([]models.ExpenseDTO, 0) // Initialize to empty slice, not nil
		for _, exp := range expenses {
			responseData = append(responseData, ToExpenseDTO(exp))
		}
		c.JSON(http.StatusOK, responseData)
		return
	}

	if filter == "all" {
		monthlyMap := make(map[string][]models.ExpenseDTO)
		distributionMap := make(map[string]float64)
		currentMonth := now.Format("January 2006")

		for _, exp := range expenses {
			month := exp.Date.Format("January 2006")
			dto := ToExpenseDTO(exp)
			monthlyMap[month] = append(monthlyMap[month], dto)

			if month == currentMonth {
				distributionMap[exp.Category] += exp.Amount
			}
		}

		// Convert monthlyMap to sorted slice
		var monthlyGroups []models.MonthlyExpenseGroup
		var months []string
		for m := range monthlyMap {
			months = append(months, m)
		}
		sort.Slice(months, func(i, j int) bool {
			t1, _ := time.Parse("January 2006", months[i])
			t2, _ := time.Parse("January 2006", months[j])
			return t1.After(t2)
		})

		for _, m := range months {
			monthlyGroups = append(monthlyGroups, models.MonthlyExpenseGroup{
				Month:    m,
				Expenses: monthlyMap[m],
			})
		}

		var currentMonthDistribution []models.CategoryDistribution
		for cat, total := range distributionMap {
			currentMonthDistribution = append(currentMonthDistribution, models.CategoryDistribution{
				Category:    cat,
				TotalAmount: total,
			})
		}

		c.JSON(http.StatusOK, models.ExpenseResponseData{
			MonthlyExpenses:          monthlyGroups,
			CurrentMonthDistribution: currentMonthDistribution,
		})
		return
	}

	var responseData []models.ExpenseDTO
	for _, exp := range expenses {
		responseData = append(responseData, ToExpenseDTO(exp))
	}

	c.JSON(http.StatusOK, responseData)
}

// GetAnalytics handles POST /analytics (intended for Insights)
func GetAnalytics(c *gin.Context) {
	userID := ExtractUserID(c)

	var request struct {
		Interval string `json:"interval"` // daily, monthly, yearly
		WalletID string `json:"wallet_id"`
		Category string `json:"category"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		// Fallback to defaults if not a POST or JSON body missing
		request.Interval = "monthly"
	}

	var startDate time.Time
	now := time.Now().UTC()

	switch request.Interval {
	case "daily":
		startDate = now.AddDate(0, 0, -7)
	case "monthly":
		startDate = now.AddDate(-1, 0, 0)
	case "yearly":
		startDate = now.AddDate(-5, 0, 0)
	default:
		startDate = now.AddDate(0, -1, 0)
	}

	query := config.DB.Where("user_id = ? AND date >= ?", userID, startDate)

	if request.WalletID != "" {
		wID, _ := strconv.ParseUint(request.WalletID, 10, 32)
		query = query.Where("wallet_id = ?", uint(wID))
	}
	if request.Category != "" {
		query = query.Where("category = ?", request.Category)
	}

	var expenses []models.Expense
	if err := query.Order("amount desc").Find(&expenses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch analytics"})
		return
	}

	// Calculate Top 5 - initialize as empty slice to avoid null in JSON
	topTransactions := make([]models.ExpenseDTO, 0)
	topN := 5
	if len(expenses) < topN {
		topN = len(expenses)
	}
	for i := 0; i < topN; i++ {
		topTransactions = append(topTransactions, ToExpenseDTO(expenses[i]))
	}

	// Calculate Chart Data
	chartMap := make(map[string]float64)
	for _, exp := range expenses {
		var label string
		switch request.Interval {
		case "daily":
			label = exp.Date.Format("Mon")
		case "monthly":
			label = exp.Date.Format("Jan")
		case "yearly":
			label = exp.Date.Format("2006")
		}
		chartMap[label] += exp.Amount
	}

	chartData := make([]models.BarChartDataPoint, 0) // Initialize as empty slice to avoid null in JSON
	// For simplicity, we just iterate the map. Chronological order would be better.
	for label, amount := range chartMap {
		chartData = append(chartData, models.BarChartDataPoint{
			ID:     label,
			Label:  label,
			Amount: amount,
		})
	}

	c.JSON(http.StatusOK, models.InsightsResponse{
		ChartData:       chartData,
		TopTransactions: topTransactions,
	})
}
