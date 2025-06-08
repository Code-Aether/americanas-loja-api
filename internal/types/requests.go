package types

import "github.com/Code-Aether/americanas-loja-api/internal/models"

// Auth Types
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100" example:"João Silva"`
	Email    string `json:"email" validate:"required,email" example:"joao@teste.com"`
	Password string `json:"password" validate:"required,min=6" example:"123456"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"joao@teste.com"`
	Password string `json:"password" validate:"required" example:"123456"`
}

type AuthResponse struct {
	Token string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  models.User `json:"user"`
}

// Product Types
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=200" example:"iPhone 15 Pro Max"`
	Description string  `json:"description" example:"Smartphone Apple com 256GB de armazenamento"`
	Price       float64 `json:"price" validate:"required,gt=0" example:"8999.99"`
	Stock       int     `json:"stock" validate:"required,gte=0" example:"50"`
	Category    string  `json:"category" validate:"required,min=2,max=100" example:"Eletrônicos"`
	SKU         string  `json:"sku" validate:"required,min=3,max=50" example:"IPHONE-15-PRO-MAX-256"`
	ImageURL    string  `json:"image_url" example:"https://example.com/iphone15.jpg"`
}

type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty" validate:"omitempty,min=2,max=200" example:"iPhone 15 Pro Max - Atualizado"`
	Description *string  `json:"description,omitempty" example:"Descrição atualizada do produto"`
	Price       *float64 `json:"price,omitempty" validate:"omitempty,gt=0" example:"8499.99"`
	Stock       *int     `json:"stock,omitempty" validate:"omitempty,gte=0" example:"45"`
	Category    *string  `json:"category,omitempty" validate:"omitempty,min=2,max=100" example:"Smartphones"`
	ImageURL    *string  `json:"image_url,omitempty" example:"https://example.com/new-iphone15.jpg"`
	SKU         *string  `json:"sku" validate:"required,min=3,max=50" example:"IPHONE-15-PRO-MAX-256"`
	Active      *bool    `json:"active,omitempty" example:"true"`
}

type ProductListResponse struct {
	Products []models.Product `json:"products"`
	Total    int64            `json:"total" example:"150"`
	Page     int              `json:"page" example:"1"`
	Limit    int              `json:"limit" example:"10"`
}
