// internal/handlers/product_handler_test.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"github.com/Code-Aether/americanas-loja-api/internal/repository"
	"github.com/Code-Aether/americanas-loja-api/internal/services"
	"github.com/Code-Aether/americanas-loja-api/internal/testutils"
	"github.com/Code-Aether/americanas-loja-api/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductHandler_GetProducts(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	redis := testutils.SetupTestRedis(t)
	productRepo := repository.NewProductRepository(db)
	productService := services.NewProductService(productRepo, redis)
	productHandler := NewProductHandler(productService)

	t.Run("✅ Listar produtos com sucesso", func(t *testing.T) {
		// Criar alguns produtos de teste
		products := []*models.Product{
			{Name: "Produto 1", SKU: "PROD-001", Price: 100.00, Stock: 10, Category: "Cat1", Active: true},
			{Name: "Produto 2", SKU: "PROD-002", Price: 200.00, Stock: 20, Category: "Cat2", Active: true},
			{Name: "Produto 3", SKU: "PROD-003", Price: 300.00, Stock: 30, Category: "Cat1", Active: false},
		}

		for _, product := range products {
			err := db.Create(product).Error
			require.NoError(t, err)
		}

		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("GET", "/products", nil)
		require.NoError(t, err)
		c.Request = req

		productHandler.GetProducts(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code, "Status deve ser 200 OK")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool), "Success deve ser true")

		data := response["data"].(map[string]interface{})
		productList := data["products"].([]interface{})
		assert.Len(t, productList, 3, "Deve retornar 3 produtos")

		assert.Equal(t, float64(3), data["total"], "Total deve ser 3")
		assert.Equal(t, float64(1), data["page"], "Page deve ser 1 (default)")
		assert.Equal(t, float64(10), data["limit"], "Limit deve ser 10 (default)")
	})

	t.Run("✅ Listar produtos com query parameters", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("GET", "/products?page=2&limit=5&category=Eletrônicos&search=iPhone", nil)
		require.NoError(t, err)
		c.Request = req

		// Mock dos query parameters
		c.Request.URL.RawQuery = "page=2&limit=5&category=Eletrônicos&search=iPhone"

		productHandler.GetProducts(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code, "Status deve ser 200 OK")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, float64(2), data["page"], "Page deve ser 2")
		assert.Equal(t, float64(5), data["limit"], "Limit deve ser 5")
	})

	t.Run("✅ Listar produtos quando não há produtos", func(t *testing.T) {
		// Usar banco limpo
		cleanDB := testutils.SetupTestDB(t)
		cleanRedis := testutils.SetupTestRedis(t)
		cleanRepo := repository.NewProductRepository(cleanDB)
		cleanService := services.NewProductService(cleanRepo, cleanRedis)
		cleanHandler := NewProductHandler(cleanService)

		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("GET", "/products", nil)
		require.NoError(t, err)
		c.Request = req

		cleanHandler.GetProducts(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code, "Status deve ser 200 OK")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		productList := data["products"].([]interface{})
		assert.Empty(t, productList, "Lista deve estar vazia")
		assert.Equal(t, float64(0), data["total"], "Total deve ser 0")
	})
}

func TestProductHandler_GetProduct(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	redis := testutils.SetupTestRedis(t)
	productRepo := repository.NewProductRepository(db)
	productService := services.NewProductService(productRepo, redis)
	productHandler := NewProductHandler(productService)

	// Criar produto de teste
	testProduct := testutils.CreateTestProduct(t, db)

	t.Run("✅ Obter produto existente", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("GET", fmt.Sprintf("/products/%d", testProduct.ID), nil)
		require.NoError(t, err)
		c.Request = req

		// Mock do parâmetro da URL
		c.Params = gin.Params{
			gin.Param{Key: "id", Value: strconv.Itoa(int(testProduct.ID))},
		}

		productHandler.GetProduct(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code, "Status deve ser 200 OK")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool), "Success deve ser true")

		product := response["data"].(map[string]interface{})
		assert.Equal(t, float64(testProduct.ID), product["id"], "ID deve estar correto")
		assert.Equal(t, testProduct.Name, product["name"], "Nome deve estar correto")
		assert.Equal(t, testProduct.SKU, product["sku"], "SKU deve estar correto")
		assert.Equal(t, testProduct.Price, product["price"], "Preço deve estar correto")
	})

	t.Run("❌ Obter produto inexistente", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("GET", "/products/99999", nil)
		require.NoError(t, err)
		c.Request = req

		c.Params = gin.Params{
			gin.Param{Key: "id", Value: "99999"},
		}

		productHandler.GetProduct(c)

		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code, "Status deve ser 404 Not Found")
		testutils.AssertErrorResponse(t, w, http.StatusNotFound)
	})

	t.Run("❌ Obter produto com ID inválido", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("GET", "/products/abc", nil)
		require.NoError(t, err)
		c.Request = req

		c.Params = gin.Params{
			gin.Param{Key: "id", Value: "abc"},
		}

		productHandler.GetProduct(c)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code, "Status deve ser 400 Bad Request")
		testutils.AssertErrorResponse(t, w, http.StatusBadRequest)
	})
}

func TestProductHandler_CreateProduct(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	redis := testutils.SetupTestRedis(t)
	productRepo := repository.NewProductRepository(db)
	productService := services.NewProductService(productRepo, redis)
	productHandler := NewProductHandler(productService)

	t.Run("✅ Criar produto com sucesso", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		// Criar usuário e adicionar ao contexto
		user := testutils.CreateTestUser(t, db)
		testutils.MockUserInContext(c, user)

		productData := types.CreateProductRequest{
			Name:        "iPhone 15 Pro Max",
			Description: "Smartphone Apple",
			Price:       8999.99,
			Stock:       50,
			Category:    "Eletrônicos",
			SKU:         "IPHONE-15-PRO-MAX",
			ImageURL:    "https://example.com/iphone15.jpg",
		}

		req, err := testutils.MockJSONRequest("POST", "/products", productData)
		require.NoError(t, err)
		c.Request = req

		productHandler.CreateProduct(c)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code, "Status deve ser 201 Created")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool), "Success deve ser true")
		assert.Contains(t, response["message"], "sucesso", "Message deve conter 'sucesso'")

		product := response["data"].(map[string]interface{})
		assert.Equal(t, productData.Name, product["name"], "Nome deve estar correto")
		assert.Equal(t, productData.Price, product["price"], "Preço deve estar correto")
		assert.Equal(t, productData.SKU, product["sku"], "SKU deve estar correto")
		assert.True(t, product["active"].(bool), "Produto deve estar ativo")
		assert.NotZero(t, product["id"], "ID deve ser gerado")

		// Verificar se foi salvo no banco
		var savedProduct models.Product
		err = db.Where("sku = ?", productData.SKU).First(&savedProduct).Error
		assert.NoError(t, err, "Produto deve estar salvo no banco")
	})

	t.Run("❌ Criar produto sem autenticação", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		// Não adicionar usuário ao contexto

		productData := types.CreateProductRequest{
			Name:  "Produto Teste",
			Price: 99.99,
			Stock: 10,
			SKU:   "TEST-001",
		}

		req, err := testutils.MockJSONRequest("POST", "/products", productData)
		require.NoError(t, err)
		c.Request = req

		productHandler.CreateProduct(c)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Status deve ser 401 Unauthorized")
		testutils.AssertErrorResponse(t, w, http.StatusUnauthorized)
	})

	t.Run("❌ Criar produto com dados inválidos", func(t *testing.T) {
		user := testutils.CreateTestUser(t, db)

		tests := []struct {
			name string
			data types.CreateProductRequest
		}{
			{
				name: "Nome vazio",
				data: types.CreateProductRequest{
					Name:  "",
					Price: 99.99,
					Stock: 10,
					SKU:   "TEST-002",
				},
			},
			{
				name: "Preço negativo",
				data: types.CreateProductRequest{
					Name:  "Produto Teste",
					Price: -10.00,
					Stock: 10,
					SKU:   "TEST-003",
				},
			},
			{
				name: "Stock negativo",
				data: types.CreateProductRequest{
					Name:  "Produto Teste",
					Price: 99.99,
					Stock: -5,
					SKU:   "TEST-004",
				},
			},
			{
				name: "SKU vazio",
				data: types.CreateProductRequest{
					Name:  "Produto Teste",
					Price: 99.99,
					Stock: 10,
					SKU:   "",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c, w := testutils.MockGinContext()

				testutils.MockUserInContext(c, user)

				req, err := testutils.MockJSONRequest("POST", "/products", tt.data)
				require.NoError(t, err)
				c.Request = req

				productHandler.CreateProduct(c)

				assert.Equal(t, http.StatusBadRequest, w.Code, "Status deve ser 400 Bad Request")
				testutils.AssertErrorResponse(t, w, http.StatusBadRequest)
			})
		}
	})

	t.Run("❌ Criar produto com SKU duplicado", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		user := testutils.CreateTestUser(t, db)
		testutils.MockUserInContext(c, user)

		// Criar produto existente
		testutils.CreateTestProduct(t, db)

		productData := types.CreateProductRequest{
			Name:  "Produto Duplicado",
			Price: 99.99,
			Stock: 10,
			SKU:   "TEST-001", // SKU já existe
		}

		req, err := testutils.MockJSONRequest("POST", "/products", productData)
		require.NoError(t, err)
		c.Request = req

		productHandler.CreateProduct(c)

		// Assertions
		assert.Equal(t, http.StatusConflict, w.Code, "Status deve ser 409 Conflict")
		testutils.AssertErrorResponse(t, w, http.StatusConflict)
	})
}

func TestProductHandler_UpdateProduct(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	redis := testutils.SetupTestRedis(t)
	productRepo := repository.NewProductRepository(db)
	productService := services.NewProductService(productRepo, redis)
	productHandler := NewProductHandler(productService)

	// Criar produto de teste
	testProduct := testutils.CreateTestProduct(t, db)

	t.Run("✅ Atualizar produto com sucesso", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		user := testutils.CreateTestUser(t, db)
		testutils.MockUserInContext(c, user)

		newName := "Nome Atualizado"
		newPrice := 299.99
		updateData := types.UpdateProductRequest{
			Name:  &newName,
			Price: &newPrice,
		}

		req, err := testutils.MockJSONRequest("PUT", fmt.Sprintf("/products/%d", testProduct.ID), updateData)
		require.NoError(t, err)
		c.Request = req

		c.Params = gin.Params{
			gin.Param{Key: "id", Value: strconv.Itoa(int(testProduct.ID))},
		}

		productHandler.UpdateProduct(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code, "Status deve ser 200 OK")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool), "Success deve ser true")

		product := response["data"].(map[string]interface{})
		assert.Equal(t, newName, product["name"], "Nome deve estar atualizado")
		assert.Equal(t, newPrice, product["price"], "Preço deve estar atualizado")

		// Verificar no banco
		var updatedProduct models.Product
		err = db.First(&updatedProduct, testProduct.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, newName, updatedProduct.Name)
		assert.Equal(t, newPrice, updatedProduct.Price)
	})

	t.Run("❌ Atualizar produto inexistente", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		user := testutils.CreateTestUser(t, db)
		testutils.MockUserInContext(c, user)

		name := "Nome Teste"
		updateData := types.UpdateProductRequest{
			Name: &name,
		}

		req, err := testutils.MockJSONRequest("PUT", "/products/99999", updateData)
		require.NoError(t, err)
		c.Request = req

		c.Params = gin.Params{
			gin.Param{Key: "id", Value: "99999"},
		}

		productHandler.UpdateProduct(c)

		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code, "Status deve ser 404 Not Found")
		testutils.AssertErrorResponse(t, w, http.StatusNotFound)
	})

	t.Run("❌ Atualizar produto sem autenticação", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		name := "Nome Teste"
		updateData := types.UpdateProductRequest{
			Name: &name,
		}

		req, err := testutils.MockJSONRequest("PUT", fmt.Sprintf("/products/%d", testProduct.ID), updateData)
		require.NoError(t, err)
		c.Request = req

		c.Params = gin.Params{
			gin.Param{Key: "id", Value: strconv.Itoa(int(testProduct.ID))},
		}

		productHandler.UpdateProduct(c)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Status deve ser 401 Unauthorized")
		testutils.AssertErrorResponse(t, w, http.StatusUnauthorized)
	})
}

func TestProductHandler_DeleteProduct(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	redis := testutils.SetupTestRedis(t)
	productRepo := repository.NewProductRepository(db)
	productService := services.NewProductService(productRepo, redis)
	productHandler := NewProductHandler(productService)

	t.Run("✅ Deletar produto com sucesso", func(t *testing.T) {
		// Criar produto para deletar
		productToDelete := testutils.CreateTestProduct(t, db)

		c, w := testutils.MockGinContext()

		admin := testutils.CreateTestAdmin(t, db)
		testutils.MockUserInContext(c, admin)

		req, err := http.NewRequest("DELETE", fmt.Sprintf("/products/%d", productToDelete.ID), nil)
		require.NoError(t, err)
		c.Request = req

		c.Params = gin.Params{
			gin.Param{Key: "id", Value: strconv.Itoa(int(productToDelete.ID))},
		}

		productHandler.DeleteProduct(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code, "Status deve ser 200 OK")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool), "Success deve ser true")

		var deletedProduct models.Product
		err = db.First(&deletedProduct, productToDelete.ID).Error
		assert.Error(t, err, "Produto deletado não deve ser encontrado")
	})

	t.Run("❌ Deletar produto inexistente", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		admin := testutils.CreateTestAdmin(t, db)
		testutils.MockUserInContext(c, admin)

		req, err := http.NewRequest("DELETE", "/products/99999", nil)
		require.NoError(t, err)
		c.Request = req

		c.Params = gin.Params{
			gin.Param{Key: "id", Value: "99999"},
		}

		productHandler.DeleteProduct(c)

		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code, "Status deve ser 404 Not Found")
		testutils.AssertErrorResponse(t, w, http.StatusNotFound)
	})

	t.Run("❌ Deletar produto sem autenticação", func(t *testing.T) {
		product := testutils.CreateTestProduct(t, db)

		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("DELETE", fmt.Sprintf("/products/%d", product.ID), nil)
		require.NoError(t, err)
		c.Request = req

		c.Params = gin.Params{
			gin.Param{Key: "id", Value: strconv.Itoa(int(product.ID))},
		}

		productHandler.DeleteProduct(c)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Status deve ser 401 Unauthorized")
		testutils.AssertErrorResponse(t, w, http.StatusUnauthorized)
	})
}
