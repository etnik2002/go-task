package main

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Password string `gorm:"not null" json:"-"`
	IsAdmin  bool   `gorm:"default:false" json:"is_admin"`
}

type Item struct {
	gorm.Model
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	LastRestock time.Time `json:"last_restock"`
}

type RestockHistory struct {
	gorm.Model
	ItemID    uint      `gorm:"not null" json:"item_id"`
	Amount    int       `gorm:"not null" json:"amount"`
	Timestamp time.Time `gorm:"not null" json:"timestamp"`
}

type RestockRequest struct {
	Amount int `json:"amount" binding:"required,min=10,max=1000"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
