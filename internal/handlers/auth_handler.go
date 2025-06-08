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
// @Param        user body types.RegisterRequest true "Dados do usuário"
// @Success      201  {object} utils.Response{data=types.AuthResponse} "Usuário criado com sucesso"
// @Failure      400  {object} utils.Response "Dados inválidos"
// @Failure      409  {object} utils.Response "Email já existe"
// @Failure      500  {object} utils.Response "Erro interno"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req types.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_DATA", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_DATA", err)
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
		if err.Error() == "user already exsists with this email" {
			utils.ErrorResponse(c, http.StatusConflict, "EMAIL_ALREADY_EXISTS", err)
			return
		}
		utils.InternalServerErrorResponse(c, err.Error(), nil)
		return
	}

	response := types.AuthResponse{
		Token: *token,
		User:  *user,
	}
	utils.SuccessResponse(c, "USER_CREATED", response)
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
		utils.BadRequestResponse(c, "INVALID_DATA", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.BadRequestResponse(c, "INVALID_DATA", err)
		return
	}

	token, user, err := h.authService.Login(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_CREDENCIALS", err)
		return
	}

	response := types.AuthResponse{
		Token: *token,
		User:  *user,
	}
	utils.SuccessResponse(c, "LOGIN_SUCCESS_MSG", response)
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
		utils.UnathorizedResponse(c, "USER_NOT_AUTHENTICATED")
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
		utils.BadRequestResponse(c, "INVALID_DATA", err)
		return
	}

	if len(req.NewPassword) < 6 {
		utils.BadRequestResponse(c, "NEW_PASSWORD_FAILED_MSG", nil)
		return
	}

	user, exists := c.Get("user")
	if !exists {
		utils.UnathorizedResponse(c, "USER_NOT_AUTHENTICATED")
		return
	}

	userModel := user.(*models.User)

	err := h.authService.ChangePassword(userModel.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		utils.BadRequestResponse(c, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "PASSWORD_CHANGE_SUCCESS_MSG", nil)
}

// RefreshToken godoc
// @Summary      Renovar token
// @Description  Gera um novo JWT token válido
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object} utils.Response{data=types.AuthResponse} "Token renovado com sucesso"
// @Failure      401  {object} utils.Response "Token inválido"
// @Failure      500  {object} utils.Response "Erro interno"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.UnathorizedResponse(c, "NO_TOKEN_REQ_MSG")
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		utils.UnathorizedResponse(c, "WRONG_TOKEN_FORMAT_MSG")
		return
	}

	newToken, user, err := h.authService.RefreshToken(tokenString)
	if err != nil {
		utils.UnathorizedResponse(c, "INVALID_TOKEN_MSG")
		return
	}

	response := types.AuthResponse{
		Token: newToken,
		User:  *user,
	}

	utils.SuccessResponse(c, "TOKEN_REFRESH_SUCCESS_MSG", response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	utils.SuccessResponse(c, "LOGOUT_SUCCESS_MSG", gin.H{
		"message": "REMOVE_TOKEN",
	})
}
