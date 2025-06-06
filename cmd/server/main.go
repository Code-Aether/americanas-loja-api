package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	
	"github.com/Code-Aether/americanas-loja-api/internal/config"
	"github.com/Code-Aether/americanas-loja-api/internal/handlers"
	"github.com/Code-Aether/americanas-loja-api/internal/repository"
	"github.com/Code-Aether/americanas-loja-api/internal/services"
	"github.com/Code-Aether/americanas-loja-api/pkg/database"
	"github.com/Code-Aether/americanas-loja-api/pkg/cache"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	db,err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	rdb := cache.NewRedisClient(cfg.RedisURL)

	productRepo := repository.NewProductRepository(db)
	userRepo    := repository.NewUserRepository(db)
	productService := services.NewProductService(productRepo, rdb)
	authService := services.NewAuthService(userRepo)

	productHandler := handlers.NewProductHandler(productService)
	authHandler := handlers.NewAuthHandler(authService)

	r := gin.Default()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	setupRoutes(r, productHandler, authHandler)

	port := os.GetEnv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting at port %s", port)
	log.Fatal(r.Run(":" + port))
}

func setupRoutes(r *gin.Engine, productHandler *handlers.ProductHandler, authHandler *handlers.AuthHandler) {
	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "OK", "service": "americanas-loja-api"})
		})

		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)

		protected := api.Group("/")
		{
			protected.GET("/products", productHandler.GetProducts)
			protected.POST("/products", productHandler.CreateProduct)
			protected.GET("/products/:id", productHandler.GetProduct)
			protected.PUT("/products/:id", productHandler.UpdateProduct)
			protected.DELETE("/products/:id", productHandler.DeleteProduct)
		}
	}
}
