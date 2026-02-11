package models

import (
	"fmt"

	"gorm.io/gorm"
)

type Wallet struct {
	gorm.Model
	Name     string  `json:"name" binding:"required"`
	Currency string  `json:"currency" binding:"required"`
	Balance  float64 `json:"balance"`
	Type     string  `json:"type"` // Cash, Debit Card, Credit Card
	CardNo   string  `json:"card_no"`
	Expiry   string  `json:"expiry"`
	Bank     string  `json:"bank"`
	Company  string  `json:"company"` // Visa, Mastercard, etc.
	Color    string  `json:"color"`
	UserID   uint    `json:"user_id"`
	User     User    `json:"-"`
}

type WalletDTO struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	Type     string  `json:"type"`
	CardNo   string  `json:"card_no"`
	Expiry   string  `json:"expiry"`
	Bank     string  `json:"bank"`
	Company  string  `json:"company"`
	Color    string  `json:"color"`
}

func (w *Wallet) ToDTO() WalletDTO {
	return WalletDTO{
		ID:       fmt.Sprintf("%d", w.ID),
		Name:     w.Name,
		Currency: w.Currency,
		Balance:  w.Balance,
		Type:     w.Type,
		CardNo:   w.CardNo,
		Expiry:   w.Expiry,
		Bank:     w.Bank,
		Company:  w.Company,
		Color:    w.Color,
	}
}
