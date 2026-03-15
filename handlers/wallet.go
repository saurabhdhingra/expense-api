package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"expense-api/config"
	"expense-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateWallet handles POST /wallets
func CreateWallet(c *gin.Context) {
	var wallet models.Wallet
	if err := c.ShouldBindJSON(&wallet); err != nil {
		log.Printf("CreateWallet Binding Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	wallet.UserID = ExtractUserID(c)

	if err := config.DB.Create(&wallet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create wallet"})
		return
	}

	c.JSON(http.StatusCreated, wallet.ToDTO())
}

// ListWallets handles GET /wallets
func ListWallets(c *gin.Context) {
	userID := ExtractUserID(c)
	var wallets []models.Wallet

	if err := config.DB.Where("user_id = ?", userID).Find(&wallets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch wallets"})
		return
	}

	var responseData []models.WalletDTO
	for _, w := range wallets {
		responseData = append(responseData, w.ToDTO())
	}

	c.JSON(http.StatusOK, responseData)
}

// UpdateWallet handles PUT /wallets/:id
func UpdateWallet(c *gin.Context) {
	walletIDStr := c.Param("id")
	walletID, err := strconv.ParseUint(walletIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Wallet ID"})
		return
	}

	userID := ExtractUserID(c)

	var wallet models.Wallet
	if err := config.DB.Where("id = ? AND user_id = ?", uint(walletID), userID).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "Wallet not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := config.DB.Model(&wallet).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update wallet"})
		return
	}

	config.DB.First(&wallet, wallet.ID)
	c.JSON(http.StatusOK, wallet.ToDTO())
}

// DeleteWallet handles DELETE /wallets/:id
func DeleteWallet(c *gin.Context) {
	walletIDStr := c.Param("id")
	walletID, err := strconv.ParseUint(walletIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Wallet ID"})
		return
	}

	userID := ExtractUserID(c)

	var wallet models.Wallet
	if err := config.DB.Where("id = ? AND user_id = ?", uint(walletID), userID).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "Wallet not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if err := config.DB.Delete(&wallet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete wallet"})
		return
	}

	c.Status(http.StatusNoContent)
}
