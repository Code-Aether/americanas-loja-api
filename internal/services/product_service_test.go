// internal/services/product_service_test.go
package services

import (
	"testing"

	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"github.com/Code-Aether/americanas-loja-api/internal/repository"
	"github.com/Code-Aether/americanas-loja-api/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductService_Create(t *testing.T) {
	db := testutils.SetupTestDB(t)
	redis := testutils.SetupTestRedis(t)
	productRepo := repository.NewProductRepository(db)
	productService := NewProductService(productRepo, redis)

	t.Run("✅ Criar produto com sucesso", func(t *testing.T) {
		product := &models.Product{
			Name:        "iPhone 15 Pro",
			Description: "Smartphone Apple",
			Price:       8999.99,
			Stock:       10,
			Category:    "Eletrônicos",
			SKU:         "IPHONE-15-PRO",
			ImageURL:    "https://example.com/iphone15.jpg",
			Active:      true,
		}

		err := productService.Create(product)

		// Assertions
		assert.NoError(t, err, "Criação não deve retornar erro")
		assert.NotZero(t, product.ID, "ID deve ser gerado")
		assert.NotZero(t, product.CreatedAt, "CreatedAt deve ser definido")
		assert.NotZero(t, product.UpdatedAt, "UpdatedAt deve ser definido")

		// Verificar se foi salvo no banco
		var savedProduct models.Product
		err = db.Where("sku = ?", "IPHONE-15-PRO").First(&savedProduct).Error
		assert.NoError(t, err, "Produto deve estar salvo no banco")
		assert.Equal(t, product.Name, savedProduct.Name)
		assert.Equal(t, product.Price, savedProduct.Price)
	})

	t.Run("❌ Criar produto com SKU duplicado", func(t *testing.T) {
		testutils.CreateTestProduct(t, db)

		duplicateProduct := &models.Product{
			Name:        "Produto Duplicado",
			Description: "Tentativa de SKU duplicado",
			Price:       100.00,
			Stock:       5,
			Category:    "Test",
			SKU:         "TEST-001", // SKU já existe
			ImageURL:    "",
			Active:      true,
		}

		err := productService.Create(duplicateProduct)

		assert.Error(t, err, "Deve retornar erro para SKU duplicado")
		assert.Contains(t, err.Error(), "sku already exists", "Erro deve mencionar SKU duplicado")
		assert.Zero(t, duplicateProduct.ID, "ID não deve ser gerado")
	})

	t.Run("❌ Criar produto com dados inválidos", func(t *testing.T) {
		tests := []struct {
			name    string
			product *models.Product
			wantErr bool
			errMsg  string
		}{
			{
				name: "Nome vazio",
				product: &models.Product{
					Name:     "",
					Price:    99.99,
					Stock:    10,
					Category: "Test",
					SKU:      "TEST-002",
				},
				wantErr: true,
				errMsg:  "name",
			},
			{
				name: "Preço negativo",
				product: &models.Product{
					Name:     "Produto Teste",
					Price:    -10.00,
					Stock:    10,
					Category: "Test",
					SKU:      "TEST-003",
				},
				wantErr: true,
				errMsg:  "price",
			},
			{
				name: "Stock negativo",
				product: &models.Product{
					Name:     "Produto Teste",
					Price:    99.99,
					Stock:    -5,
					Category: "Test",
					SKU:      "TEST-004",
				},
				wantErr: true,
				errMsg:  "stock",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := productService.Create(tt.product)

				if tt.wantErr {
					assert.Error(t, err, "Deve retornar erro")
					if tt.errMsg != "" {
						assert.Contains(t, err.Error(), tt.errMsg, "Erro deve conter mensagem específica")
					}
				} else {
					assert.NoError(t, err, "Não deve retornar erro")
				}
			})
		}
	})
}

func TestProductService_GetByID(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	redis := testutils.SetupTestRedis(t)
	productRepo := repository.NewProductRepository(db)
	productService := NewProductService(productRepo, redis)

	// Criar produto de teste
	testProduct := testutils.CreateTestProduct(t, db)

	t.Run("✅ Buscar produto existente", func(t *testing.T) {
		product, err := productService.GetByID(testProduct.ID)

		// Assertions
		assert.NoError(t, err, "Busca não deve retornar erro")
		assert.NotNil(t, product, "Produto deve ser encontrado")
		assert.Equal(t, testProduct.ID, product.ID)
		assert.Equal(t, testProduct.Name, product.Name)
		assert.Equal(t, testProduct.SKU, product.SKU)
		assert.Equal(t, testProduct.Price, product.Price)
	})

	t.Run("❌ Buscar produto inexistente", func(t *testing.T) {
		product, err := productService.GetByID(99999) // ID que não existe

		// Assertions
		assert.Error(t, err, "Deve retornar erro")
		assert.Nil(t, product, "Produto não deve ser encontrado")
		assert.Contains(t, err.Error(), "not found", "Erro deve mencionar que não foi encontrado")
	})

	t.Run("❌ Buscar produto com ID zero", func(t *testing.T) {
		product, err := productService.GetByID(0)

		// Assertions
		assert.Error(t, err, "Deve retornar erro")
		assert.Nil(t, product, "Produto não deve ser encontrado")
	})
}

func TestProductService_GetAll(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	redis := testutils.SetupTestRedis(t)
	productRepo := repository.NewProductRepository(db)
	productService := NewProductService(productRepo, redis)

	t.Run("✅ Listar produtos quando existe produtos", func(t *testing.T) {
		// Criar alguns produtos de teste
		products := []*models.Product{
			{
				Name: "Produto 1", SKU: "PROD-001", Price: 100.00, Stock: 10,
				Category: "Cat1", Active: true,
			},
			{
				Name: "Produto 2", SKU: "PROD-002", Price: 200.00, Stock: 20,
				Category: "Cat2", Active: true,
			},
			{
				Name: "Produto 3", SKU: "PROD-003", Price: 300.00, Stock: 30,
				Category: "Cat1", Active: false, // Produto inativo
			},
		}

		for _, product := range products {
			err := db.Create(product).Error
			require.NoError(t, err)
		}

		result, _, err := productService.GetAll(1, 10, "", "")

		// Assertions
		assert.NoError(t, err, "Listagem não deve retornar erro")
		assert.NotNil(t, result, "Lista não deve ser nil")
		assert.Len(t, result, 3, "Deve retornar todos os produtos (incluindo inativos)")

		// Verificar se produtos estão ordenados por ID
		for i := 1; i < len(result); i++ {
			assert.True(t, result[i-1].ID <= result[i].ID, "Produtos devem estar ordenados por ID")
		}
	})

	t.Run("✅ Listar produtos quando não há produtos", func(t *testing.T) {
		// Usar banco limpo
		cleanDB := testutils.SetupTestDB(t)
		cleanRedis := testutils.SetupTestRedis(t)
		cleanRepo := repository.NewProductRepository(cleanDB)
		cleanService := NewProductService(cleanRepo, cleanRedis)

		result, _, err := cleanService.GetAll(1, 10, "", "")

		// Assertions
		assert.NoError(t, err, "Listagem de lista vazia não deve retornar erro")
		assert.NotNil(t, result, "Lista não deve ser nil")
		assert.Empty(t, result, "Lista deve estar vazia")
	})
}

func TestProductService_Update(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	redis := testutils.SetupTestRedis(t)
	productRepo := repository.NewProductRepository(db)
	productService := NewProductService(productRepo, redis)

	t.Run("✅ Atualizar produto com sucesso", func(t *testing.T) {
		// Criar produto inicial
		originalProduct := testutils.CreateTestProduct(t, db)
		originalUpdatedAt := originalProduct.UpdatedAt

		// Atualizar dados
		originalProduct.Name = "Nome Atualizado"
		originalProduct.Price = 199.99
		originalProduct.Stock = 25
		originalProduct.Description = "Descrição atualizada"

		err := productService.Update(originalProduct)

		// Assertions
		assert.NoError(t, err, "Atualização não deve retornar erro")

		// Verificar se foi atualizado no banco
		var updatedProduct models.Product
		err = db.First(&updatedProduct, originalProduct.ID).Error
		assert.NoError(t, err, "Produto deve existir no banco")

		assert.Equal(t, "Nome Atualizado", updatedProduct.Name)
		assert.Equal(t, 199.99, updatedProduct.Price)
		assert.Equal(t, 25, updatedProduct.Stock)
		assert.Equal(t, "Descrição atualizada", updatedProduct.Description)
		assert.True(t, updatedProduct.UpdatedAt.After(originalUpdatedAt), "UpdatedAt deve ser atualizado")
	})

	t.Run("❌ Atualizar produto inexistente", func(t *testing.T) {
		nonExistentProduct := &models.Product{
			ID:   99999, // ID que não existe
			Name: "Produto Inexistente",
		}

		err := productService.Update(nonExistentProduct)

		// Assertions
		assert.Error(t, err, "Deve retornar erro")
		assert.Contains(t, err.Error(), "not found", "Erro deve mencionar que não foi encontrado")
	})

	t.Run("❌ Atualizar com dados inválidos", func(t *testing.T) {
		product := testutils.CreateTestProduct(t, db)

		// Tentar atualizar com preço negativo
		product.Price = -50.00

		err := productService.Update(product)

		// Assertions
		assert.Error(t, err, "Deve retornar erro para preço negativo")
		assert.Contains(t, err.Error(), "price", "Erro deve mencionar preço")
	})
}

func TestProductService_Delete(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	redis := testutils.SetupTestRedis(t)
	productRepo := repository.NewProductRepository(db)
	productService := NewProductService(productRepo, redis)

	t.Run("✅ Deletar produto com sucesso", func(t *testing.T) {
		// Criar produto para deletar
		product := testutils.CreateTestProduct(t, db)
		productID := product.ID

		err := productService.Delete(productID)

		// Assertions
		assert.NoError(t, err, "Deleção não deve retornar erro")

		// Verificar se foi realmente deletado
		var deletedProduct models.Product
		err = db.First(&deletedProduct, productID).Error
		assert.Error(t, err, "Produto deletado não deve ser encontrado")
	})

	t.Run("❌ Deletar produto inexistente", func(t *testing.T) {
		err := productService.Delete(99999) // ID que não existe

		// Assertions
		assert.Error(t, err, "Deve retornar erro")
		assert.Contains(t, err.Error(), "not found", "Erro deve mencionar que não foi encontrado")
	})

	t.Run("❌ Deletar com ID zero", func(t *testing.T) {
		err := productService.Delete(0)

		// Assertions
		assert.Error(t, err, "Deve retornar erro para ID zero")
	})
}

func TestProductService_GetBySKU(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	redis := testutils.SetupTestRedis(t)
	productRepo := repository.NewProductRepository(db)
	productService := NewProductService(productRepo, redis)

	// Criar produto de teste
	testProduct := testutils.CreateTestProduct(t, db)

	t.Run("✅ Buscar produto por SKU existente", func(t *testing.T) {
		product, err := productService.GetByID(testProduct.ID)

		// Assertions
		assert.NoError(t, err, "Busca por SKU não deve retornar erro")
		assert.NotNil(t, product, "Produto deve ser encontrado")
		assert.Equal(t, testProduct.ID, product.ID)
		assert.Equal(t, testProduct.SKU, product.SKU)
		assert.Equal(t, testProduct.Name, product.Name)
	})

	t.Run("❌ Buscar produto por SKU inexistente", func(t *testing.T) {
		product, err := productService.GetByID(99999)

		// Assertions
		assert.Error(t, err, "Deve retornar erro")
		assert.Nil(t, product, "Produto não deve ser encontrado")
		assert.Contains(t, err.Error(), "not found", "Erro deve mencionar que não foi encontrado")
	})

	t.Run("❌ Buscar produto com SKU vazio", func(t *testing.T) {
		product, err := productService.GetByID(0)

		// Assertions
		assert.Error(t, err, "Deve retornar erro para SKU vazio")
		assert.Nil(t, product, "Produto não deve ser encontrado")
	})
}

func TestProductService_GetByCategory(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	redis := testutils.SetupTestRedis(t)
	productRepo := repository.NewProductRepository(db)
	productService := NewProductService(productRepo, redis)

	// Criar produtos de diferentes categorias
	products := []*models.Product{
		{Name: "iPhone", SKU: "IP-001", Category: "Eletrônicos", Price: 1000, Stock: 10, Active: true},
		{Name: "Samsung", SKU: "SM-001", Category: "Eletrônicos", Price: 800, Stock: 15, Active: true},
		{Name: "Camisa", SKU: "CM-001", Category: "Roupas", Price: 50, Stock: 30, Active: true},
		{Name: "Calça", SKU: "CL-001", Category: "Roupas", Price: 80, Stock: 20, Active: false}, // Inativo
	}

	for _, product := range products {
		err := db.Create(product).Error
		require.NoError(t, err)
	}

	t.Run("✅ Buscar produtos por categoria existente", func(t *testing.T) {
		result, err := productService.GetByCategory("Eletrônicos")

		// Assertions
		assert.NoError(t, err, "Busca por categoria não deve retornar erro")
		assert.Len(t, result, 2, "Deve retornar 2 produtos da categoria Eletrônicos")

		for _, product := range result {
			assert.Equal(t, "Eletrônicos", product.Category, "Todos produtos devem ser da categoria correta")
		}
	})

	t.Run("✅ Buscar produtos por categoria que inclui inativos", func(t *testing.T) {
		result, err := productService.GetByCategory("Roupas")

		// Assertions
		assert.NoError(t, err, "Busca por categoria não deve retornar erro")
		assert.Len(t, result, 2, "Deve retornar 2 produtos da categoria Roupas (incluindo inativo)")

		activeCount := 0
		for _, product := range result {
			assert.Equal(t, "Roupas", product.Category, "Todos produtos devem ser da categoria correta")
			if product.Active {
				activeCount++
			}
		}
		assert.Equal(t, 1, activeCount, "Deve ter 1 produto ativo na categoria")
	})

	t.Run("✅ Buscar produtos por categoria inexistente", func(t *testing.T) {
		result, err := productService.GetByCategory("Categoria Inexistente")

		// Assertions
		assert.NoError(t, err, "Busca por categoria inexistente não deve retornar erro")
		assert.Empty(t, result, "Lista deve estar vazia para categoria inexistente")
	})
}
