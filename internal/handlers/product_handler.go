package handlers

import (
	"strconv"

	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"github.com/Code-Aether/americanas-loja-api/internal/services"
	"github.com/Code-Aether/americanas-loja-api/pkg/utils"
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
	var req models.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "INVALID_DATA", err)
		return
	}

	if req.Name == "" {
		utils.BadRequestResponse(c, "NAME_IS_REQUIRED", nil)
		return
	}

	if req.Price <= 0 {
		utils.BadRequestResponse(c, "PRICE_NEEDS_TO_BE_BIGGER_THAN_ZERO", nil)
		return
	}

	if req.Stock < 0 {
		utils.BadRequestResponse(c, "STOCK_CANNOT_BE_NEGATIVE", nil)
		return
	}

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		ImageURL:    req.ImageURL,
		Active:      true,
	}

	if err := h.productService.Create(product); err != nil {
		utils.InternalServerErrorResponse(c, "ERROR_CREATING_PRODUCT", err)
		return
	}

	utils.SuccessResponse(c, "PRODUCT_CREATED_WITH_SUCCESS", product)
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	category := c.Query("category")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 10
	}

	products, total, err := h.productService.GetAll(page, limit, category, search)
	if err != nil {
		utils.InternalServerErrorResponse(c, "SEARCH_PRODUCTS_ERROR", err)
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	pagination := utils.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}
	utils.PaginatedSuccessResponse(c, "PRODUCTS_LISTED_SUCCESS", products, pagination)
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "INVALID_ID", err)
		return
	}

	product, err := h.productService.GetByID(uint(id))
	if err != nil {
		utils.NotFoundResponse(c, "PRODUCT_NOT_FOUND", err)
	}

	utils.SuccessResponse(c, "PRODUCT_FOUND", product)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "INVALID_ID", err)
		return
	}

	var req models.ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "INVALID_DATA", err)
		return
	}

	product, err := h.productService.GetByID(uint(id))
	if err != nil {
		utils.NotFoundResponse(c, "PRODUCT_NOT_FOUND", err)
		return
	}

	if req.Name != nil {
		if *req.Name == "" {
			utils.BadRequestResponse(c, "NAME_IS_REQUIRED", nil)
			return
		}
		product.Name = *req.Name
	}

	if req.Description != nil {
		product.Description = *req.Description
	}

	if req.Price != nil {
		if *req.Price <= 0 {
			utils.BadRequestResponse(c, "PRICE_NEEDS_TO_BE_BIGGER_THAN_ZERO", nil)
			return
		}
		product.Price = *req.Price
	}

	if req.Stock != nil {
		if *req.Stock < 0 {
			utils.BadRequestResponse(c, "STOCK_CANNOT_BE_NEGATIVE", nil)
			return
		}

		product.Stock = *req.Stock
	}

	if req.Category != nil {
		product.Category = *req.Category
	}

	if req.ImageURL != nil {
		product.ImageURL = *req.ImageURL
	}

	if req.Active != nil {
		product.Active = *req.Active
	}

	if err := h.productService.Update(product); err != nil {
		utils.InternalServerErrorResponse(c, "UPDATE_PRODUCT_ERROR", err)
		return
	}

	utils.SuccessResponse(c, "PRODUCT_UPDATE_SUCCEFULL", product)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "INVALID_ID", err)
		return
	}

	if err := h.productService.Delete(uint(id)); err != nil {
		utils.InternalServerErrorResponse(c, "ERROR_DELETING_PRODUCT", err)
		return
	}

	utils.SuccessResponse(c, "PRODUCT_DELETED_WITH_SUCCESS", nil)
}
