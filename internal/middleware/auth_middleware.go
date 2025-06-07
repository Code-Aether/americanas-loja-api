package middleware

import (
	"fmt"
	"log"
	"strings"

	"github.com/Code-Aether/americanas-loja-api/internal/services"
	"github.com/Code-Aether/americanas-loja-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddlewareLog("Verifying autentication for: %s %s", c.Request.Method, c.Request.URL.Path)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			authMiddlewareLog("Did not receive a token")
			utils.UnathorizedResponse(c, "NO_TOKEN_REQ_MSG")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			authMiddlewareLog("Invalid Header format for token, should use Bearer <token>")
			utils.UnathorizedResponse(c, "WRONG_TOKEN_FORMAT_MSG"+" Use: Bearer <token>")
			c.Abort()
		}

		user, err := m.authService.GetUserByToken(tokenString)
		if err != nil {
			authMiddlewareLog("Token is invalid, or expired")
			utils.UnathorizedResponse(c, "INVALID_TOKEN_OR_EXPIRED")
			c.Abort()
			return
		}

		authMiddlewareLog("User %s (role: %s) authenticated", user, user.Role)

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)

		c.Next()

		authMiddlewareLog("Request has processed for %s", user.Email)
	}
}

func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.RequireAuth()(c)

		if c.IsAborted() {
			return
		}

		userRole, exists := c.Get("user_role")
		if !exists {
			utils.InternalServerErrorResponse(c, "INTERNAL_ERROR_MSG", nil)
			c.Abort()
			return
		}

		if userRole.(string) != "admin" {
			utils.UnathorizedResponse(c, "ACCESS_DENIED"+"ONLY_FOR_ADMINS")
			c.Abort()
			return
		}

		c.Next()
	}
}

// Don't abort if not authenticated
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.Next()
			return
		}

		user, err := m.authService.GetUserByToken(tokenString)
		if err == nil {
			c.Set("user", user)
			c.Set("user_id", user.ID)
			c.Set("user_role", user.Role)
		}

		c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		m.RequireAuth()(c)

		if c.IsAborted() {
			return
		}

		userRole, exists := c.Get("user_role")
		if !exists {
			utils.InternalServerErrorResponse(c, "INTERNAL_ERROR", nil)
			c.Abort()
			return
		}

		if userRole.(string) != requiredRole {
			utils.UnathorizedResponse(c, "ACCESS_DENIED"+"NEED_ROLE"+": "+requiredRole)
			c.Abort()
			return
		}

		c.Next()
	}
}

func authMiddlewareLog(format string, v ...any) {
	logPrefix := "[AUTH-MIDDLEWARE]"
	message := fmt.Sprintf(format, v)
	log.Printf("%s %s", logPrefix, message)
}
