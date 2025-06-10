package services

import (
	"testing"
	"time"

	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"github.com/Code-Aether/americanas-loja-api/internal/repository"
	"github.com/Code-Aether/americanas-loja-api/internal/testutils"
	"github.com/Code-Aether/americanas-loja-api/internal/types"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	t.Run("🧪 Registro com sucesso", func(t *testing.T) {
		user := &models.User{
			Name:     "John Doe",
			Email:    "john@test.com",
			Password: "password123",
			Role:     "user",
			Active:   true,
		}

		token, err := authService.Register(user)

		// Assertions
		assert.NoError(t, err, "Registro não deve retornar erro")
		assert.NotEmpty(t, token, "Token deve ser gerado")
		assert.NotEqual(t, "password123", user.Password, "Password deve estar hashada")
		assert.NotZero(t, user.ID, "User ID deve ser gerado")

		// Verificar se password foi hashada corretamente
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123"))
		assert.NoError(t, err, "Password hash deve ser válida")

		// Verificar se foi salvo no banco
		var savedUser models.User
		err = db.Where("email = ?", "john@test.com").First(&savedUser).Error
		assert.NoError(t, err, "Usuário deve estar salvo no banco")
		assert.Equal(t, user.Name, savedUser.Name)
		assert.Equal(t, user.Email, savedUser.Email)
	})

	t.Run("❌ Registro com email duplicado", func(t *testing.T) {
		// Criar primeiro usuário
		testutils.CreateTestUser(t, db)

		// Tentar criar segundo usuário com mesmo email
		user := &models.User{
			Name:     "Another User",
			Email:    "test@test.com", // Email já existe
			Password: "password123",
			Role:     "user",
			Active:   true,
		}

		token, err := authService.Register(user)

		// Assertions
		assert.Error(t, err, "Deve retornar erro para email duplicado")
		assert.Empty(t, token, "Token não deve ser gerado")
		assert.Contains(t, err.Error(), "user already exists", "Erro deve mencionar email duplicado")
	})

	t.Run("❌ Registro com dados inválidos", func(t *testing.T) {
		tests := []struct {
			name    string
			user    *models.User
			wantErr bool
			errMsg  string
		}{
			{
				name: "Email vazio",
				user: &models.User{
					Name:     "Test User",
					Email:    "",
					Password: "password123",
				},
				wantErr: true,
				errMsg:  "email is required",
			},
			{
				name: "Password muito curta",
				user: &models.User{
					Name:     "Test User",
					Email:    "test2@test.com",
					Password: "123", // Muito curta
				},
				wantErr: true,
				errMsg:  "password",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				token, err := authService.Register(tt.user)

				if tt.wantErr {
					assert.Error(t, err, "Deve retornar erro")
					assert.Empty(t, token, "Token não deve ser gerado")
					if tt.errMsg != "" {
						assert.Contains(t, err.Error(), tt.errMsg, "Erro deve conter mensagem específica")
					}
				} else {
					assert.NoError(t, err, "Não deve retornar erro")
					assert.NotEmpty(t, token, "Token deve ser gerado")
				}
			})
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	// Criar usuário de teste
	user := testutils.CreateTestUser(t, db)

	t.Run("✅ Login com sucesso", func(t *testing.T) {
		token, returnedUser, err := authService.Login(types.LoginRequest{
			Email:    user.Email,
			Password: "password123",
		})

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotNil(t, returnedUser)
		assert.Equal(t, user.Email, returnedUser.Email)
	})

	t.Run("❌ Login com email inexistente", func(t *testing.T) {
		token, returnedUser, err := authService.Login(types.LoginRequest{
			Email:    "nonexistent@test.com",
			Password: "password123",
		})

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, returnedUser)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("❌ Login com password incorreta", func(t *testing.T) {
		token, returnedUser, err := authService.Login(types.LoginRequest{
			Email:    user.Email,
			Password: "wrongpassword",
		})

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, returnedUser)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("❌ Login com usuário inativo", func(t *testing.T) {
		// Criar usuário inativo
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		require.NoError(t, err)

		inactiveUser := &models.User{
			Name:     "Inactive User",
			Email:    "inactive@test.com",
			Password: string(hashedPassword),
			Role:     "user",
			Active:   false,
		}
		err = userRepo.Create(inactiveUser)
		require.NoError(t, err)

		// Explicitly update the user to ensure it is inactive in the DB
		err = db.Model(inactiveUser).Update("active", false).Error
		assert.NoError(t, err, "Erro ao atualizar usuário inativo")

		token, returnedUser, err := authService.Login(types.LoginRequest{
			Email:    inactiveUser.Email,
			Password: "password123",
		})

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, returnedUser)
		assert.Equal(t, "user is inactive", err.Error())
	})
}

func TestAuthService_GenerateJWT(t *testing.T) {
	// Setup
	authService := &AuthService{
		jwtSecret: "test-secret-key",
	}

	user := &models.User{
		ID:    123,
		Email: "test@jwt.com",
		Role:  "user",
	}

	t.Run("✅ Gerar JWT com sucesso", func(t *testing.T) {
		token, _, err := authService.GenerateJWT(user)

		// Assertions
		assert.NoError(t, err, "Geração de JWT não deve retornar erro")
		assert.NotEmpty(t, token, "Token deve ser gerado")

		// Verificar se o token é válido
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})

		assert.NoError(t, err, "Token deve ser válido")
		assert.True(t, parsedToken.Valid, "Token deve estar válido")

		// Verificar claims
		if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
			assert.Equal(t, float64(user.ID), claims["user_id"], "User ID deve estar correto")
			assert.Equal(t, user.Email, claims["email"], "Email deve estar correto")
			assert.Equal(t, user.Role, claims["role"], "Role deve estar correto")
			assert.Equal(t, "americanas-loja-api", claims["iss"], "Issuer deve estar correto")
		} else {
			t.Fatal("Claims não puderam ser lidas")
		}
	})

	t.Run("⏰ Verificar expiração do token", func(t *testing.T) {
		token, _, err := authService.GenerateJWT(user)
		require.NoError(t, err)

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})
		require.NoError(t, err)

		claims := parsedToken.Claims.(jwt.MapClaims)
		exp := int64(claims["exp"].(float64))

		// Verificar se expira em 24 horas (aproximadamente)
		expectedExp := time.Now().Add(24 * time.Hour).Unix()
		assert.InDelta(t, expectedExp, exp, 60, "Token deve expirar em ~24 horas")
	})
}

func TestAuthService_GetUserByToken(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	// Criar usuário de teste
	user := testutils.CreateTestUser(t, db)

	t.Run("✅ Buscar usuário por token válido", func(t *testing.T) {
		// Gerar token válido
		token, _, err := authService.GenerateJWT(user)
		require.NoError(t, err)

		// Buscar usuário pelo token
		foundUser, err := authService.GetUserByToken(token)

		// Assertions
		assert.NoError(t, err, "Busca por token válido não deve retornar erro")
		assert.NotNil(t, foundUser, "Usuário deve ser encontrado")
		assert.Equal(t, user.ID, foundUser.ID)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.Role, foundUser.Role)
	})

	t.Run("❌ Token inválido", func(t *testing.T) {
		invalidToken := "token.invalido.aqui"

		foundUser, err := authService.GetUserByToken(invalidToken)

		// Assertions
		assert.Error(t, err, "Token inválido deve retornar erro")
		assert.Nil(t, foundUser, "Usuário não deve ser encontrado")
	})

	t.Run("❌ Token expirado", func(t *testing.T) {
		// Criar token com expiração no passado
		claims := jwt.MapClaims{
			"user_id": user.ID,
			"email":   user.Email,
			"role":    user.Role,
			"exp":     time.Now().Add(-1 * time.Hour).Unix(), // Expirado há 1 hora
			"iat":     time.Now().Add(-2 * time.Hour).Unix(),
			"iss":     "americanas-loja-api",
			"sub":     user.Email,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		require.NoError(t, err)

		foundUser, err := authService.GetUserByToken(tokenString)

		// Assertions
		assert.Error(t, err, "Token expirado deve retornar erro")
		assert.Nil(t, foundUser, "Usuário não deve ser encontrado")
		assert.Contains(t, err.Error(), "token is expired", "Erro deve mencionar token expirado")
	})

	t.Run("❌ Usuário não existe mais", func(t *testing.T) {
		// Criar token para usuário que será deletado
		deletedUser := &models.User{
			Name:     "To Be Deleted",
			Email:    "deleted@test.com",
			Password: "hashed_password",
			Role:     "user",
			Active:   true,
		}
		db.Create(deletedUser)

		// Gerar token
		token, _, err := authService.GenerateJWT(deletedUser)
		require.NoError(t, err)

		// Deletar usuário
		db.Delete(deletedUser)

		// Tentar buscar com token
		foundUser, err := authService.GetUserByToken(token)

		// Assertions
		assert.Error(t, err, "Token de usuário deletado deve retornar erro")
		assert.Nil(t, foundUser, "Usuário não deve ser encontrado")
	})
}

func TestAuthService_HashPassword(t *testing.T) {
	authService := &AuthService{}

	t.Run("✅ Hash password com sucesso", func(t *testing.T) {
		password := "password123"

		hashedPassword, err := authService.hashPassword(password)

		// Assertions
		assert.NoError(t, err, "Hash não deve retornar erro")
		assert.NotEmpty(t, hashedPassword, "Hash não deve estar vazio")
		assert.NotEqual(t, password, hashedPassword, "Hash deve ser diferente da senha original")

		// Verificar se o hash é válido
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		assert.NoError(t, err, "Hash deve ser válido")
	})

	t.Run("✅ Hashes diferentes para mesma senha", func(t *testing.T) {
		password := "samepassword"

		hash1, err1 := authService.hashPassword(password)
		hash2, err2 := authService.hashPassword(password)

		// Assertions
		assert.NoError(t, err1, "Primeiro hash não deve retornar erro")
		assert.NoError(t, err2, "Segundo hash não deve retornar erro")
		assert.NotEqual(t, hash1, hash2, "Hashes devem ser diferentes (salt diferente)")

		// Ambos devem ser válidos
		err1 = bcrypt.CompareHashAndPassword([]byte(hash1), []byte(password))
		err2 = bcrypt.CompareHashAndPassword([]byte(hash2), []byte(password))
		assert.NoError(t, err1, "Primeiro hash deve ser válido")
		assert.NoError(t, err2, "Segundo hash deve ser válido")
	})
}
