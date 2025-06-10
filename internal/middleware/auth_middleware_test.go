// internal/middleware/auth_middleware_test.go
package middleware

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Code-Aether/americanas-loja-api/internal/repository"
	"github.com/Code-Aether/americanas-loja-api/internal/services"
	"github.com/Code-Aether/americanas-loja-api/internal/testutils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, "test-secret")
	authMiddleware := NewAuthMiddleware(authService)

	t.Run("✅ Autenticação com token válido", func(t *testing.T) {
		c, w := testutils.MockGinContext()
		// Criar usuário de teste
		testUser := testutils.CreateTestUser(t, db)

		// Gerar token para o usuário
		token, _, err := authService.GenerateJWT(testUser)
		require.NoError(t, err)

		req, err := http.NewRequest("GET", "/protected", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		c.Request = req

		nextCalled := false
		authMiddleware.RequireAuth()(c)
		if !c.IsAborted() {
			nextCalled = true
		}

		assert.True(t, nextCalled, "Next() deve ser chamado")
		assert.Equal(t, http.StatusOK, w.Code, "Status deve ser 200 OK")
	})

	t.Run("❌ Sem Authorization header", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("GET", "/protected", nil)
		require.NoError(t, err)
		c.Request = req

		authMiddleware.RequireAuth()(c)

		assert.True(t, c.IsAborted(), "Request deve ser abortado")
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Status deve ser 401 Unauthorized")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"].(string), "authorization", "Message deve mencionar authorization")
	})

	t.Run("❌ Token sem formato Bearer", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("GET", "/protected", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "invalid-token")
		c.Request = req

		authMiddleware.RequireAuth()(c)

		assert.True(t, c.IsAborted(), "Request deve ser abortado")
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Status deve ser 401 Unauthorized")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"].(string), "invalid token format", "Message deve mencionar formato inválido")
	})

	t.Run("❌ Token inválido", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("GET", "/protected", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalid-token")
		c.Request = req

		authMiddleware.RequireAuth()(c)

		assert.True(t, c.IsAborted(), "Request deve ser abortado")
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Status deve ser 401 Unauthorized")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"].(string), "invalid token", "Message deve mencionar token invalid")
	})

	t.Run("❌ Token de usuário inexistente", func(t *testing.T) {

		c, w := testutils.MockGinContext()
		// Criar token para usuário inexistente
		claims := services.JWTClaims{
			UserID: 999,
			Email:  "nonexistent@test.com",
			Role:   "user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				Issuer:    "americanas-loja-api",
				Subject:   "nonexistent@test.com",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		require.NoError(t, err)

		req, err := http.NewRequest("GET", "/protected", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		c.Request = req

		authMiddleware.RequireAuth()(c)

		assert.True(t, c.IsAborted(), "Request deve ser abortado")
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Status deve ser 401 Unauthorized")
	})
}

func TestAuthMiddleware_RequireAdmin(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, "test-secret")
	authMiddleware := NewAuthMiddleware(authService)

	// Criar usuários de teste
	regularUser := testutils.CreateTestUser(t, db)
	adminUser := testutils.CreateTestAdmin(t, db)

	t.Run("✅ Acesso com usuário admin", func(t *testing.T) {
		c, w := testutils.MockGinContext()

		token, _, err := authService.GenerateJWT(adminUser)
		require.NoError(t, err)

		req, err := http.NewRequest("GET", "/admin", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		c.Request = req

		nextCalled := false
		c.Set("next", func() {
			nextCalled = true
		})

		authMiddleware.RequireAdmin()(c)

		assert.True(t, nextCalled, "Next() deve ser chamado")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("❌ Acesso com usuário comum", func(t *testing.T) {

		c, w := testutils.MockGinContext()

		token, _, err := authService.GenerateJWT(regularUser)
		require.NoError(t, err)

		req, err := http.NewRequest("GET", "/admin", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		c.Request = req

		authMiddleware.RequireAdmin()(c)

		assert.True(t, c.IsAborted(), "Context deve estar abortado")
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("❌ Acesso sem token", func(t *testing.T) {

		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("GET", "/admin", nil)
		require.NoError(t, err)
		c.Request = req

		nextCalled := false
		c.Set("next", func() {
			nextCalled = true
		})

		authMiddleware.RequireAdmin()(c)

		assert.False(t, nextCalled, "Next() não deve ser chamado sem token")
		assert.True(t, c.IsAborted(), "Context deve estar abortado")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthMiddleware_OptionalAuth(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, "test-secret")
	authMiddleware := NewAuthMiddleware(authService)

	user := testutils.CreateTestUser(t, db)

	t.Run("✅ Acesso com token válido", func(t *testing.T) {

		c, w := testutils.MockGinContext()

		token, _, err := authService.GenerateJWT(user)
		require.NoError(t, err)

		req, err := http.NewRequest("GET", "/public", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		c.Request = req

		nextCalled := false
		c.Set("next", func() {
			nextCalled = true
		})

		authMiddleware.OptionalAuth()(c)

		assert.True(t, nextCalled, "Next() deve ser chamado")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("✅ Acesso sem token (permitido)", func(t *testing.T) {

		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("GET", "/public", nil)
		require.NoError(t, err)
		c.Request = req

		nextCalled := false
		c.Set("next", func() {
			nextCalled = true
		})

		authMiddleware.OptionalAuth()(c)

		assert.True(t, nextCalled, "Next() deve ser chamado")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("✅ Acesso com token inválido (continua sem auth)", func(t *testing.T) {

		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("GET", "/public", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalid-token")
		c.Request = req

		nextCalled := false
		c.Set("next", func() {
			nextCalled = true
		})

		authMiddleware.OptionalAuth()(c)

		assert.True(t, nextCalled, "Next() deve ser chamado")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("✅ Acesso com formato de token inválido (continua sem auth)", func(t *testing.T) {

		c, w := testutils.MockGinContext()

		req, err := http.NewRequest("GET", "/public", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "invalid-token") // Sem "Bearer "
		c.Request = req

		nextCalled := false
		c.Set("next", func() {
			nextCalled = true
		})

		authMiddleware.OptionalAuth()(c)

		assert.True(t, nextCalled, "Next() deve ser chamado")
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAuthMiddleware_Integration(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, "test-secret")
	authMiddleware := NewAuthMiddleware(authService)

	user := testutils.CreateTestUser(t, db)
	admin := testutils.CreateTestAdmin(t, db)

	t.Run("🔄 Fluxo completo: User -> RequireAuth -> RequireAdmin", func(t *testing.T) {
		userToken, _, err := authService.GenerateJWT(user)
		require.NoError(t, err)

		// 1. RequireAuth com usuário comum (deve passar)
		c1, w1 := testutils.MockGinContext()
		req1, _ := http.NewRequest("GET", "/protected", nil)
		req1.Header.Set("Authorization", "Bearer "+userToken)
		c1.Request = req1

		next1Called := false
		c1.Set("next", func() {
			next1Called = true
		})

		authMiddleware.RequireAuth()(c1)
		assert.True(t, next1Called, "RequireAuth deve passar para usuário comum")
		assert.Equal(t, http.StatusOK, w1.Code)

		// 2. RequireAdmin com usuário comum (deve falhar)
		c2, w2 := testutils.MockGinContext()
		req2, _ := http.NewRequest("GET", "/admin", nil)
		req2.Header.Set("Authorization", "Bearer "+userToken)
		c2.Request = req2

		authMiddleware.RequireAdmin()(c2)
		assert.True(t, c2.IsAborted(), "Context deve estar abortado")
		assert.Equal(t, http.StatusForbidden, w2.Code)

		// 3. RequireAdmin com admin (deve passar)
		adminToken, _, err := authService.GenerateJWT(admin)
		require.NoError(t, err)

		c3, w3 := testutils.MockGinContext()
		req3, _ := http.NewRequest("GET", "/admin", nil)
		req3.Header.Set("Authorization", "Bearer "+adminToken)
		c3.Request = req3

		authMiddleware.RequireAdmin()(c3)
		assert.True(t, !c3.IsAborted(), "RequireAdmin deve passar para admin")
		assert.Equal(t, http.StatusOK, w3.Code)
	})
}
