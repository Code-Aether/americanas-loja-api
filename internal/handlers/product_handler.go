package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"github.com/Code-Aether/americanas-loja-api/internal/services"
	"github.com/Code-Aether/americanas-loja-api/internal/types"
	"github.com/Code-Aether/americanas-loja-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ProductHandler struct {
	productService *services.ProductService
	validator      *validator.Validate
}

func NewProductHandler(productService *services.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		validator:      validator.New(),
	}
}

// CreateProduct godoc
// @Summary      Criar novo produto
// @Description  Cria um novo produto no sistema (requer autenticação)
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        product body types.CreateProductRequest true "Dados do produto"
// @Success      201 {object} utils.Response{data=models.Product} "Produto criado com sucesso"
// @Failure      400 {object} utils.Response "Dados inválidos"
// @Failure      401 {object} utils.Response "Token inválido"
// @Failure      409 {object} utils.Response "SKU já existe"
// @Failure      500 {object} utils.Response "Erro interno"
// @Router       /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req types.CreateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "INVALID_DATA", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.BadRequestResponse(c, "INVALID_DATA", err)
		return
	}

	user, err := checkUserLogged(c)
	if err != nil {
		utils.UnathorizedResponse(c, "USER_NOT_AUTHENTICATED")
		return
	}

	productHandlerLog("User %s (role: %s) creating product: %s",
		user.Email, user.Role, req.Name)

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		SKU:         req.SKU,
		Category:    req.Category,
		ImageURL:    req.ImageURL,
		Active:      true,
	}

	if err := h.productService.Create(product); err != nil {
		if err.Error() == "sku already exists" {
			utils.ErrorResponse(c, http.StatusConflict, "SKU_ALREADY_EXISTS", err)
			return
		}
		utils.InternalServerErrorResponse(c, "ERROR_CREATING_PRODUCT", err)
		return
	}

	utils.SuccessResponse(c, "PRODUCT_CREATED_WITH_SUCCESS", product)
}

// GetProducts godoc
// @Summary      Listar produtos
// @Description  Retorna lista paginada de produtos disponíveis
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        page     query int    false "Número da página" default(1)
// @Param        limit    query int    false "Itens por página" default(10)
// @Param        category query string false "Filtrar por categoria" example("Eletrônicos")
// @Param        search   query string false "Buscar produtos" example("iPhone")
// @Success      200 {object} utils.Response{data=types.ProductListResponse} "Lista de produtos"
// @Failure      500 {object} utils.Response "Erro interno"
// @Router       /products [get]
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

// GetProduct godoc
// @Summary      Obter produto específico
// @Description  Retorna detalhes de um produto pelo ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id path int true "ID do produto" example(1)
// @Success      200 {object} utils.Response{data=models.Product} "Produto encontrado"
// @Failure      400 {object} utils.Response "ID inválido"
// @Failure      404 {object} utils.Response "Produto não encontrado"
// @Failure      500 {object} utils.Response "Erro interno"
// @Router       /products/{id} [get]
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

// UpdateProduct godoc
// @Summary      Atualizar produto
// @Description  Atualiza um produto existente (requer autenticação)
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "ID do produto" example(1)
// @Param        product body types.UpdateProductRequest true "Dados para atualização"
// @Success      200 {object} utils.Response{data=models.Product} "Produto atualizado com sucesso"
// @Failure      400 {object} utils.Response "Dados inválidos"
// @Failure      401 {object} utils.Response "Token inválido"
// @Failure      404 {object} utils.Response "Produto não encontrado"
// @Failure      500 {object} utils.Response "Erro interno"
// @Router       /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "INVALID_ID", err)
		return
	}

	var req types.UpdateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "INVALID_DATA", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.BadRequestResponse(c, "INVALID_DATA", err)
		return
	}

	user, err := checkUserLogged(c)
	if err != nil {
		utils.UnathorizedResponse(c, "USER_NOT_AUTHENTICATED")
		return
	}

	productHandlerLog("User %s (role %s) updating product ID: %d",
		user.Email, user.Role, id)

	product, err := h.productService.GetByID(uint(id))
	if err != nil {
		utils.NotFoundResponse(c, "PRODUCT_NOT_FOUND", err)
		return
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
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

// DeleteProduct godoc
// @Summary      Deletar produto
// @Description  Remove um produto do sistema (apenas admins)
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "ID do produto" example(1)
// @Success      200 {object} utils.Response "Produto deletado com sucesso"
// @Failure      400 {object} utils.Response "ID inválido"
// @Failure      401 {object} utils.Response "Token inválido"
// @Failure      403 {object} utils.Response "Acesso negado"
// @Failure      404 {object} utils.Response "Produto não encontrado"
// @Failure      500 {object} utils.Response "Erro interno"
// @Router       /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "INVALID_ID", err)
		return
	}

	user, err := checkUserLogged(c)
	if err != nil {
		utils.BadRequestResponse(c, err.Error(), err)
		return
	}

	productHandlerLog("Admin %s deletando produto ID: %d", user.Email, id)

	_, err = h.productService.GetByID(uint(id))
	if err != nil {
		utils.NotFoundResponse(c, "PRODUCT_NOT_FOUND", err)
		return
	}

	if err := h.productService.Delete(uint(id)); err != nil {
		utils.InternalServerErrorResponse(c, "ERROR_DELETING_PRODUCT", err)
		return
	}

	utils.SuccessResponse(c, "PRODUCT_DELETED_WITH_SUCCESS", nil)
}

func checkUserLogged(c *gin.Context) (*models.User, error) {
	user, exists := c.Get("user")
	if !exists {
		return nil, fmt.Errorf("USER_NOT_AUTHENTICATED")
	}
	userModel := user.(*models.User)
	return userModel, nil
}

func productHandlerLog(format string, v ...any) {
	prefix := "[PRODUCT_HANDLER]"
	message := fmt.Sprintf(format, v)
	log.Printf("%s %s", prefix, message)
}
