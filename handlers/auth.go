package handlers

import (
	"expense-api/config"
	"expense-api/models"
	"expense-api/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RegisterUser handles POST /register
func RegisterUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 1. Check for unique email
	var existingUser models.User
	if config.DB.Where("email = ?", user.Email).First(&existingUser).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"message": "Email already registered"})
		return
	}

	// 2. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	// 3. Save User
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	// 4. Generate Token and respond
	token, _ := utils.GenerateToken(user.ID)
	c.JSON(http.StatusCreated, gin.H{"token": token})
}

// LoginUser handles POST /login
func LoginUser(c *gin.Context) {
	var loginRequest struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	log.Printf("Login Request: Email=%s, Password=%s", loginRequest.Email, loginRequest.Password)

	// 1. Find User by Email
	var user models.User
	if err := config.DB.Where("email = ?", loginRequest.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found with email: " + loginRequest.Email})
		return
	}
	log.Printf("User retrieved: ID=%d, Hashed Password=%s", user.ID, user.Password)
	log.Printf("Password from login request: %s", loginRequest.Password)

	// 2. Verify Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Password Incorrect, err : " + err.Error()})
		return
	}

	// 3. Generate Token and respond
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetProfile handles GET /me
func GetProfile(c *gin.Context) {
	userID := ExtractUserID(c)
	var user models.User

	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, models.PublicUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	})
}

// RefreshToken handles POST /refreshToken
func RefreshToken(c *gin.Context) {
	var request struct {
		OldToken string `json:"oldToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// For simplicity, we just generate a new token if we can extract a user ID from the old one,
	// even if the old one is expired (as long as it was signed with our secret).
	claims, _ := utils.ValidateToken(request.OldToken)
	if claims == nil || claims.UserID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or malformed token"})
		return
	}

	// If we got claims, use the UserID
	newToken, err := utils.GenerateToken(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"newToken":   newToken,
		"expiryTime": 86400, // 24 hours in seconds
	})
}

func ExtractUserID(c *gin.Context) uint {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(uint)
	}
	return 0
}
