package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/Code-Aether/americanas-loja-api/internal/config"
	"github.com/Code-Aether/americanas-loja-api/internal/handlers"
	"github.com/Code-Aether/americanas-loja-api/internal/middleware"
	"github.com/Code-Aether/americanas-loja-api/internal/repository"
	"github.com/Code-Aether/americanas-loja-api/internal/services"
	"github.com/Code-Aether/americanas-loja-api/pkg/cache"
	"github.com/Code-Aether/americanas-loja-api/pkg/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	db, err := database.NewConnection(cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	rdb := cache.NewRedisClient(cfg.RedisURL, "", 0)

	err = database.AutoMigrate(db)
	if err != nil {
		log.Fatal("failed to auto migrate:", err)
	}

	productRepo := repository.NewProductRepository(db)
	userRepo := repository.NewUserRepository(db)
	productService := services.NewProductService(productRepo, rdb)
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)

	productHandler := handlers.NewProductHandler(productService)
	authHandler := handlers.NewAuthHandler(authService)

	err = database.SeedData(db)
	if err != nil {
		log.Fatal("failed to seed data:", err)
	}

	err = database.SeedAdminUser(db)
	if err != nil {
		log.Fatal("failed to create admin user:", err)
	}

	r := gin.Default()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	setupRoutes(r, productHandler, authHandler, authService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting at http://localhost:%s", port)
	log.Fatal(r.Run(":" + port))
}

func setupRoutes(r *gin.Engine, productHandler *handlers.ProductHandler, authHandler *handlers.AuthHandler, authService *services.AuthService) {
	root := r.Group("/")
	{
		root.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "OK", "service": "americanas-loja-api"})
		})
	}

	api := r.Group("/api/v1")
	{
		authMiddleware := middleware.NewAuthMiddleware(authService)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		public := api.Group("/")
		public.Use(authMiddleware.OptionalAuth())
		{
			public.GET("/products", productHandler.GetProducts)
			public.GET("/products/:id", productHandler.GetProduct)
		}
		protected := api.Group("/")
		protected.Use(authMiddleware.RequireAuth())
		{
			protected.POST("/products", productHandler.CreateProduct)
			protected.PUT("/products/:id", productHandler.UpdateProduct)
		}

		adminProtected := api.Group("/")
		adminProtected.Use(authMiddleware.RequireAdmin())
		{
			adminProtected.DELETE("/products/:id", productHandler.DeleteProduct)
		}
	}
}
