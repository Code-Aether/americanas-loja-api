package middleware

import (
	"fmt"
	"log"
	"net/http"
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
		if c.IsAborted() {
			return
		}

		authMiddlewareLog("Verifying autentication for: %s %s", c.Request.Method, c.Request.URL.Path)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			authMiddlewareLog("Did not receive a token")
			utils.ErrorResponse(c, http.StatusUnauthorized, "authorization header is missing", nil)
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			authMiddlewareLog("Invalid Header format for token, should use Bearer <token>")
			utils.ErrorResponse(c, http.StatusUnauthorized, "invalid token format. Use: Bearer <token>", nil)
			c.Abort()
			return
		}

		user, err := m.authService.GetUserByToken(tokenString)
		if err != nil {
			authMiddlewareLog("Token is invalid, or expired")
			utils.ErrorResponse(c, http.StatusUnauthorized, "invalid token", nil)
			c.Abort()
			return
		}

		authMiddlewareLog("User %s (role: %s) authenticated", user.Email, user.Role)

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)

		if !c.IsAborted() {
			if next, exists := c.Get("next"); exists {
				next.(func())()
			}
		}

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
			utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error", nil)
			c.Abort()
			return
		}

		if userRole.(string) != "admin" {
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied. Only admin users can access this resource", nil)
			c.Abort()
			return
		}

		if !c.IsAborted() {
			if next, exists := c.Get("next"); exists {
				next.(func())()
			}
		}
	}
}

// Don't abort if not authenticated
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			if next, exists := c.Get("next"); exists {
				next.(func())()
			}
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			if next, exists := c.Get("next"); exists {
				next.(func())()
			}
			return
		}

		user, err := m.authService.GetUserByToken(tokenString)
		if err == nil {
			c.Set("user", user)
			c.Set("user_id", user.ID)
			c.Set("user_role", user.Role)
		}

		if next, exists := c.Get("next"); exists {
			next.(func())()
		}
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
			utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error", nil)
			c.Abort()
			return
		}

		if userRole.(string) != requiredRole {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Access denied. Need role: "+requiredRole, nil)
			c.Abort()
			return
		}

		if !c.IsAborted() {
			if next, exists := c.Get("next"); exists {
				next.(func())()
			}
		}
	}
}

func authMiddlewareLog(format string, v ...any) {
	logPrefix := "[AUTH-MIDDLEWARE]"
	message := fmt.Sprintf(format, v...)
	log.Printf("%s %s", logPrefix, message)
}
