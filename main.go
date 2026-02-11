package main

import (
	"log"
	"net/http"

	"expense-api/config"
	"expense-api/handlers"
	"expense-api/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDatabase()

	r := gin.Default()

	r.POST("/register", handlers.RegisterUser)
	r.POST("/login", handlers.LoginUser)

	authenticated := r.Group("/")
	authenticated.Use(middleware.AuthMiddleware())
	{
		authenticated.GET("/me", handlers.GetProfile)
		authenticated.POST("/expenses", handlers.CreateExpense)
		authenticated.PUT("/expenses/:id", handlers.UpdateExpense)
		authenticated.DELETE("/expense/:id", handlers.DeleteExpense)

		authenticated.GET("/expenses", handlers.ListExpenses)

		authenticated.POST("/wallets", handlers.CreateWallet)
		authenticated.GET("/wallets", handlers.ListWallets)
		authenticated.PUT("/wallets/:id", handlers.UpdateWallet)
		authenticated.DELETE("/wallets/:id", handlers.DeleteWallet)

		authenticated.POST("/analytics", handlers.GetAnalytics)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP", "message": "Expense Tracker API is running"})
	})

	log.Println("Server listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
