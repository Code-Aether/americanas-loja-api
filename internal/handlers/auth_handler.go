package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"github.com/Code-Aether/americanas-loja-api/internal/services"
	"github.com/Code-Aether/americanas-loja-api/internal/types"
	"github.com/Code-Aether/americanas-loja-api/pkg/utils"
)

type AuthHandler struct {
	authService *services.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

// Register godoc
// @Summary      Registrar novo usuário
// @Description  Cria uma nova conta de usuário no sistema
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user body types.RegisterRequest true "user data"
// @Success      201  {object} utils.Response{data=types.AuthResponse} "user created with success"
// @Failure      400  {object} utils.Response "invalid data"
// @Failure      409  {object} utils.Response "email already exists"
// @Failure      500  {object} utils.Response "internal error"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req types.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid data", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid data", err)
		return
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     "user",
		Active:   true,
	}

	token, err := h.authService.Register(user)
	if err != nil {
		if err.Error() == "user already exists" {
			utils.ErrorResponse(c, http.StatusConflict, "email already exists", err)
			return
		}
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response := types.AuthResponse{
		Token: *token,
		User:  *user,
	}
	utils.SuccessResponseWithStatus(c, http.StatusCreated, "user created with success", response)
}

// Login godoc
// @Summary      Fazer login
// @Description  Autentica usuário e retorna JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body types.LoginRequest true "Credenciais de login"
// @Success      200  {object} utils.Response{data=types.AuthResponse} "Login realizado com sucesso"
// @Failure      400  {object} utils.Response "Dados inválidos"
// @Failure      401  {object} utils.Response "Credenciais inválidas"
// @Failure      500  {object} utils.Response "Erro interno"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req types.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid data", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.BadRequestResponse(c, "invalid data", err)
		return
	}

	token, user, err := h.authService.Login(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "invalid user or password", err)
		return
	}

	response := types.AuthResponse{
		Token: *token,
		User:  *user,
	}
	utils.SuccessResponse(c, "login successfull", response)
}

// GetProfile godoc
// @Summary      Obter perfil do usuário
// @Description  Retorna informações do usuário autenticado
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object} utils.Response{data=models.User} "Perfil do usuário"
// @Failure      401  {object} utils.Response "Token inválido"
// @Failure      500  {object} utils.Response "Erro interno"
// @Router       /user/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "USER_NOT_AUTHENTICATED", nil)
		return
	}

	utils.SuccessResponse(c, "USER_PROFILE", user)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	type ChangePasswordRequest struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid data", err)
		return
	}

	if len(req.NewPassword) < 6 {
		utils.BadRequestResponse(c, "new password is invalid, should be bigger than 6 characters", nil)
		return
	}

	user, exists := c.Get("user")
	if !exists {
		utils.UnathorizedResponse(c, "user not authenticated")
		return
	}

	userModel := user.(*models.User)

	err := h.authService.ChangePassword(userModel.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		utils.BadRequestResponse(c, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "password change sucessfull", nil)
}

// RefreshToken godoc
// @Summary      Renovar token
// @Description  Gera um novo JWT token válido
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object} utils.Response{data=types.AuthResponse} "token renewed successfully"
// @Failure      401  {object} utils.Response "invalid token format"
// @Failure      500  {object} utils.Response "internal error"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "authorization header is required", nil)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		utils.ErrorResponse(c, http.StatusUnauthorized, "invalid token format. Use: Bearer <token>", nil)
		return
	}

	newToken, user, err := h.authService.RefreshToken(tokenString)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "invalid or expired token", err)
		return
	}

	response := types.AuthResponse{
		Token: newToken,
		User:  *user,
	}
	utils.SuccessResponse(c, "token renewed successfully", response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	utils.SuccessResponse(c, "logout successfull", gin.H{
		"message": "revoke token from user",
	})
}
