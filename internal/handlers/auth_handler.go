package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/Code-Aether/americanas-loja-api/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Register endpoint - TODO",	
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Login endpoint - TODO",
	})
}
