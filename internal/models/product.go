package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint           `json:"id" gorm:"PrimaryKey"`
	Name        string         `json:"name" gorm:"not null;size:255" validate:"required,min=2,max=255"`
	Description string         `json:"description" gorm:"type:text"`
	Price       float64        `json:"price" gorm:"not null" validate:"required,gt=0"`
	Stock       int            `json:"stock" gorm:"not null;default:0" validate:"min=0"`
	Category    string         `json:"category"`
	SKU         string         `json:"sku" gorm:"uniqueIndex;size:100"`
	Active      bool           `json:"active" gorm:"default:true"`
	ImageURL    string         `json:"image_url" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type ProductCreateRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"min=0"`
	Category    string  `json:"category"`
	ImageURL    string  `json:"image_url" gorm:"type:text"`
	SKU         string  `json:"sku"`
}

type ProductUpdateRequest struct {
	Name        *string  `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	Stock       *int     `json:"stock,omitempty" validate:"omitempty,min=0"`
	Category    *string  `json:"category_id,omitempty"`
	ImageURL    *string  `json:"image_url" gorm:"type:text"`
	Active      *bool    `json:"active,omitempty"`
	SKU         *string  `json:"sku"`
}
