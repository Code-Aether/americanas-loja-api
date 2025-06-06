package handlers

import (
	"net/http"

	"github.com/Code-Aether/americanas-loja-api/internal/services"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService *services.ProductService
}

func NewProductHandler(productService *services.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"message": "Create product - TODO",
	})
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get all products - TODO",
		"data":    []string{},
	})
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Get product by ID - TODO",
		"id":      id,
	})
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Update product - TODO",
		"id":      id,
	})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Delete product - TODO",
		"id":      id,
	})
}
