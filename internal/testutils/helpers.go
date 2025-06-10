package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Code-Aether/americanas-loja-api/internal/config"
	"github.com/Code-Aether/americanas-loja-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silenciar logs nos testes
	})
	assert.NoError(t, err, "Erro ao conectar com banco de teste")

	// Auto migrate
	err = db.AutoMigrate(&models.User{}, &models.Product{})
	assert.NoError(t, err, "Erro ao migrar banco de teste")

	return db
}

func CreateTestUser(t *testing.T, db *gorm.DB) *models.User {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	assert.NoError(t, err, "Erro ao hashear senha de teste")
	user := &models.User{
		Name:     "Test User",
		Email:    "test@test.com",
		Password: string(hash),
		Role:     "user",
		Active:   true,
	}

	err = db.Create(user).Error
	assert.NoError(t, err, "Erro ao criar usuário de teste")

	return user
}

func CreateTestAdmin(t *testing.T, db *gorm.DB) *models.User {
	admin := &models.User{
		Name:     "Test Admin",
		Email:    "admin@test.com",
		Password: "$2a$10$example_hashed_password", // Hash simulado
		Role:     "admin",
		Active:   true,
	}

	err := db.Create(admin).Error
	assert.NoError(t, err, "Erro ao criar admin de teste")

	return admin
}

func CreateTestProduct(t *testing.T, db *gorm.DB) *models.Product {
	product := &models.Product{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Stock:       10,
		Category:    "Test Category",
		SKU:         "TEST-001",
		ImageURL:    "https://test.com/image.jpg",
		Active:      true,
	}

	err := db.Create(product).Error
	assert.NoError(t, err, "Erro ao criar produto de teste")

	return product
}

func MockGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	return c, w
}

func MockJSONRequest(method, url string, body interface{}) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int) {
	assert.Equal(t, expectedCode, w.Code, "Status code incorreto")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Erro ao fazer unmarshal da resposta")

	assert.True(t, response["success"].(bool), "Response success deve ser true")
	assert.NotEmpty(t, response["message"], "Response message não pode estar vazio")
}

func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int) {
	assert.Equal(t, expectedCode, w.Code, "Status code incorreto")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Erro ao fazer unmarshal da resposta")

	assert.False(t, response["success"].(bool), "Response success deve ser false")
	assert.NotEmpty(t, response["message"], "Response message não pode estar vazio")
}

func GetTestConfig() *config.Config {
	return &config.Config{
		DBDriver:     "sqlite",
		DBSQlitePath: ":memory:",
		Port:         "8080",
		JWTSecret:    "test-secret-key-super-secure",
		RedisURL:     "redis://localhost:6379",
	}
}

func MockUserInContext(c *gin.Context, user *models.User) {
	c.Set("user", user)
	c.Set("user_id", user.ID)
	c.Set("user_role", user.Role)
}

func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expected interface{}) {
	var actual interface{}
	err := json.Unmarshal(w.Body.Bytes(), &actual)
	assert.NoError(t, err, "Erro ao fazer unmarshal da resposta")

	expectedJSON, _ := json.Marshal(expected)
	actualJSON, _ := json.Marshal(actual)

	assert.JSONEq(t, string(expectedJSON), string(actualJSON), "JSON response diferente do esperado")
}

func WaitForAsync(t *testing.T, timeout time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func SetupTestRedis(t *testing.T) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Ping Redis to check connection
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}

	return rdb
}
