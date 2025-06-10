// internal/handlers/auth_handler_test.go
package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"github.com/Code-Aether/americanas-loja-api/internal/repository"
	"github.com/Code-Aether/americanas-loja-api/internal/services"
	"github.com/Code-Aether/americanas-loja-api/internal/testutils"
	"github.com/Code-Aether/americanas-loja-api/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_Register(t *testing.T) {
	db := testutils.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, "test-secret")
	authHandler := NewAuthHandler(authService)

	t.Run("✅ Registro com sucesso", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		// Preparar request
		registerData := types.RegisterRequest{
			Name:     "John Doe",
			Email:    "john@test.com",
			Password: "password123",
		}

		req, err := testutils.MockJSONRequest("POST", "/auth/register", registerData)
		require.NoError(t, err)

		c.Request = req

		authHandler.Register(c)

		assert.Equal(t, http.StatusCreated, w.Code, "Status deve ser 201 Created")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool), "Success deve ser true")
		assert.Contains(t, response["message"], "success", "Message deve conter 'success'")

		// Verificar se retorna dados
		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["token"], "Token deve ser retornado")

		user := data["user"].(map[string]interface{})
		assert.Equal(t, registerData.Name, user["name"], "Nome deve estar correto")
		assert.Equal(t, registerData.Email, user["email"], "Email deve estar correto")
		assert.Equal(t, "user", user["role"], "Role deve ser 'user'")
		assert.True(t, user["active"].(bool), "Active deve ser true")

		var savedUser models.User
		err = db.Where("email = ?", registerData.Email).First(&savedUser).Error
		assert.NoError(t, err, "Usuário deve estar salvo no banco")
	})

	t.Run("❌ Registro com dados inválidos", func(t *testing.T) {
		tests := []struct {
			name           string
			data           types.RegisterRequest
			expectedStatus int
		}{
			{
				name: "Email vazio",
				data: types.RegisterRequest{
					Name:     "Test User",
					Email:    "",
					Password: "password123",
				},
				expectedStatus: http.StatusBadRequest,
			},
			{
				name: "Email inválido",
				data: types.RegisterRequest{
					Name:     "Test User",
					Email:    "email-invalido",
					Password: "password123",
				},
				expectedStatus: http.StatusBadRequest,
			},
			{
				name: "Password muito curta",
				data: types.RegisterRequest{
					Name:     "Test User",
					Email:    "test@test.com",
					Password: "123",
				},
				expectedStatus: http.StatusBadRequest,
			},
			{
				name: "Nome vazio",
				data: types.RegisterRequest{
					Name:     "",
					Email:    "test2@test.com",
					Password: "password123",
				},
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c, w := testutils.MockGinContext()

				req, err := testutils.MockJSONRequest("POST", "/auth/register", tt.data)
				require.NoError(t, err)
				c.Request = req

				authHandler.Register(c)

				assert.Equal(t, tt.expectedStatus, w.Code, "Status code deve estar correto")
				testutils.AssertErrorResponse(t, w, tt.expectedStatus)
			})
		}
	})

	t.Run("❌ Registro com email duplicado", func(t *testing.T) {
		testutils.CreateTestUser(t, db)

		c, w := testutils.MockGinContext()

		registerData := types.RegisterRequest{
			Name:     "Another User",
			Email:    "test@test.com",
			Password: "password123",
		}

		req, err := testutils.MockJSONRequest("POST", "/auth/register", registerData)
		require.NoError(t, err)
		c.Request = req

		authHandler.Register(c)

		assert.Equal(t, http.StatusConflict, w.Code, "Status deve ser 409 Conflict")
		testutils.AssertErrorResponse(t, w, http.StatusConflict)
	})

	t.Run("❌ Registro com JSON inválido", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("POST", "/auth/register", nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		authHandler.Register(c)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Status deve ser 400 Bad Request")
		testutils.AssertErrorResponse(t, w, http.StatusBadRequest)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, "test-secret")
	authHandler := NewAuthHandler(authService)

	// Criar usuário de teste para login
	user := &models.User{
		Name:     "Test User",
		Email:    "login@test.com",
		Password: "password123",
		Role:     "user",
		Active:   true,
	}
	_, err := authService.Register(user)
	require.NoError(t, err)

	t.Run("✅ Login com sucesso", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		loginData := types.LoginRequest{
			Email:    "login@test.com",
			Password: "password123",
		}

		req, err := testutils.MockJSONRequest("POST", "/auth/login", loginData)
		require.NoError(t, err)
		c.Request = req

		authHandler.Login(c)

		assert.Equal(t, http.StatusOK, w.Code, "Status deve ser 200 OK")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool), "Success deve ser true")
		assert.Contains(t, response["message"], "success", "Message deve conter 'success'")

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["token"], "Token deve ser retornado")

		user := data["user"].(map[string]interface{})
		assert.Equal(t, loginData.Email, user["email"], "Email deve estar correto")
		assert.Equal(t, "user", user["role"], "Role deve estar correto")
	})

	t.Run("❌ Login com credenciais inválidas", func(t *testing.T) {
		tests := []struct {
			name string
			data types.LoginRequest
		}{
			{
				name: "Email inexistente",
				data: types.LoginRequest{
					Email:    "inexistente@test.com",
					Password: "password123",
				},
			},
			{
				name: "Password incorreta",
				data: types.LoginRequest{
					Email:    "login@test.com",
					Password: "senha_errada",
				},
			},
			{
				name: "Email vazio",
				data: types.LoginRequest{
					Email:    "",
					Password: "password123",
				},
			},
			{
				name: "Password vazia",
				data: types.LoginRequest{
					Email:    "login@test.com",
					Password: "",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c, w := testutils.MockGinContext()

				req, err := testutils.MockJSONRequest("POST", "/auth/login", tt.data)
				require.NoError(t, err)
				c.Request = req

				authHandler.Login(c)

				// Para email/password vazios, deve ser 400 (validação)
				// Para credenciais incorretas, deve ser 401
				expectedStatus := http.StatusUnauthorized
				if tt.data.Email == "" || tt.data.Password == "" {
					expectedStatus = http.StatusBadRequest
				}

				assert.Equal(t, expectedStatus, w.Code, "Status code deve estar correto")
				testutils.AssertErrorResponse(t, w, expectedStatus)
			})
		}
	})
}

func TestAuthHandler_GetProfile(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, "test-secret")
	authHandler := NewAuthHandler(authService)

	t.Run("✅ Obter perfil com usuário autenticado", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		// Criar usuário de teste
		user := testutils.CreateTestUser(t, db)

		// Simular usuário no contexto (middleware já validou)
		testutils.MockUserInContext(c, user)

		authHandler.GetProfile(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code, "Status deve ser 200 OK")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool), "Success deve ser true")

		userData := response["data"].(map[string]interface{})
		assert.Equal(t, user.Email, userData["email"], "Email deve estar correto")
		assert.Equal(t, user.Name, userData["name"], "Nome deve estar correto")
		assert.Equal(t, user.Role, userData["role"], "Role deve estar correto")
	})

	t.Run("❌ Obter perfil sem usuário no contexto", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		// Não adicionar usuário no contexto (simular middleware falhando)

		authHandler.GetProfile(c)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Status deve ser 401 Unauthorized")
		testutils.AssertErrorResponse(t, w, http.StatusUnauthorized)
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, "test-secret")
	authHandler := NewAuthHandler(authService)

	t.Run("✅ Renovar token com sucesso", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		// Criar usuário de teste
		user := testutils.CreateTestUser(t, db)

		// Gerar token para o usuário
		token, _, err := authService.GenerateJWT(user)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/auth/refresh", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		c.Request = req

		authHandler.RefreshToken(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code, "Status deve ser 200 OK")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool), "Success deve ser true")

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["token"], "Novo token deve ser retornado")

		userData := data["user"].(map[string]interface{})
		assert.Equal(t, user.Email, userData["email"], "Email deve estar correto")
	})

	t.Run("❌ Renovar token sem usuário no contexto", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("POST", "/auth/refresh", nil)
		require.NoError(t, err)
		c.Request = req

		authHandler.RefreshToken(c)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Status deve ser 401 Unauthorized")
		testutils.AssertErrorResponse(t, w, http.StatusUnauthorized)
	})
}

func TestAuthHandler_Placeholders(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, "test-secret")
	authHandler := NewAuthHandler(authService)

	t.Run("📝 ChangePassword placeholder", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		user := testutils.CreateTestUser(t, db)
		testutils.MockUserInContext(c, user)

		// Mock request body
		reqBody := map[string]string{
			"old_password": "password123",
			"new_password": "newpassword123",
		}
		req, err := testutils.MockJSONRequest("POST", "/auth/change-password", reqBody)
		require.NoError(t, err)
		c.Request = req

		authHandler.ChangePassword(c)

		assert.Equal(t, http.StatusOK, w.Code, "Status deve ser 200 OK")
		testutils.AssertSuccessResponse(t, w, http.StatusOK)
	})

	t.Run("📝 Logout placeholder", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		user := testutils.CreateTestUser(t, db)
		testutils.MockUserInContext(c, user)

		authHandler.Logout(c)

		assert.Equal(t, http.StatusOK, w.Code, "Status deve ser 200 OK")
		testutils.AssertSuccessResponse(t, w, http.StatusOK)
	})
}
