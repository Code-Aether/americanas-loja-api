package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"github.com/Code-Aether/americanas-loja-api/internal/services"
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

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "INVALID_DATA", err)
		return
	}

	if len(req.Password) < 6 {
		utils.BadRequestResponse(c, "PASSWORD_MIN_CHARS_ERROR", nil)
		return
	}

	if len(req.Name) < 2 {
		utils.BadRequestResponse(c, "USER_NAME_MIN_CHARS_ERROR", nil)
		return
	}

	if err := h.validator.Var(req.Email, "required,email"); err != nil {
		utils.BadRequestResponse(c, "INVALID_EMAIL", err)
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error(), nil)
		return
	}

	user.Password = ""

	utils.SuccessResponse(c, "USER_CREATED", gin.H{
		"user":    user,
		"message": "LOGIN_ALLOWED_MSG",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "INVALID_DATA", err)
		return
	}

	if req.Email == "" {
		utils.BadRequestResponse(c, "EMAIL_IS_REQUIRED", nil)
		return
	}

	if req.Password == "" {
		utils.BadRequestResponse(c, "PASSWORD_IS_REQUIRED", nil)
		return
	}

	loginResponse, err := h.authService.Login(req)
	if err != nil {
		utils.UnathorizedResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "LOGIN_SUCCESS_MSG", loginResponse)
}

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

	newToken, err := h.authService.RefreshToken(tokenString)
	if err != nil {
		utils.UnathorizedResponse(c, "INVALID_TOKEN_MSG")
		return
	}

	utils.SuccessResponse(c, "TOKEN_REFRESH_SUCCESS_MSG", gin.H{
		"token": newToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	utils.SuccessResponse(c, "LOGOUT_SUCCESS_MSG", gin.H{
		"message": "REMOVE_TOKEN",
	})
}
