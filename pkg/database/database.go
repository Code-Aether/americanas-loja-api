package database

import (
	"fmt"
	"log"

	"github.com/Code-Aether/americanas-loja-api/internal/config"
	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"golang.org/x/crypto/bcrypt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewConnection(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	if cfg.DBDriver == "sqlite" {
		log.Println("Using SQLite database")
		dialector = sqlite.Open(cfg.DBSQlitePath)
	} else {
		log.Println("Using PostgreSQL database")
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Sao_Paulo",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
		dialector = postgres.Open(dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	log.Println("Connected to the database")
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	log.Println("Starting database migration...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Product{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migrated successfully")

	if !db.Migrator().HasTable(&models.Product{}) {
		log.Println("Error: Table products has not created!")
		return fmt.Errorf("products table does not exists")
	}

	return nil
}

func SeedData(db *gorm.DB) error {
	log.Println("Starting database seeding...")

	var count int64
	db.Model(&models.User{}).Count(&count)

	if count > 0 {
		log.Println("Database already seeded... Skipping")
		return nil
	}

	products := []models.Product{
		{
			Name:        "iPhone 15 Pro Max",
			Description: "O iPhone mais avançado da Apple com chip A17 Pro, câmera de 48MP e tela ProMotion de 6.7 polegadas.",
			Price:       8999.99,
			Stock:       25,
			Category:    "Eletrônicos",
			SKU:         "IPHONE-15-PRO-MAX-001",
			Active:      true,
			ImageURL:    "https://example.com/iphone15.jpg",
		},
		{
			Name:        "MacBook Air M3",
			Description: "Notebook ultrafino com chip M3, 8GB RAM, 256GB SSD. Perfeito para trabalho e estudos.",
			Price:       12499.99,
			Stock:       15,
			Category:    "Eletrônicos",
			SKU:         "MACBOOK-AIR-M3-002",
			Active:      true,
			ImageURL:    "https://example.com/macbook.jpg",
		},
		{
			Name:        "Smart TV LG 55\" 4K",
			Description: "Smart TV LG NanoCell 55 polegadas 4K UHD com WebOS e HDR10.",
			Price:       2799.99,
			Stock:       30,
			Category:    "Eletrônicos",
			SKU:         "LG-TV-55-4K-003",
			Active:      true,
			ImageURL:    "https://example.com/tv-lg.jpg",
		},
		{
			Name:        "PlayStation 5",
			Description: "Console de última geração da Sony com SSD ultra-rápido e controle DualSense.",
			Price:       4499.99,
			Stock:       10,
			Category:    "Games",
			SKU:         "SONY-PS5-004",
			Active:      true,
			ImageURL:    "https://example.com/ps5.jpg",
		},
		{
			Name:        "Airfryer Philips XL",
			Description: "Fritadeira elétrica sem óleo, 4L de capacidade, ideal para famílias.",
			Price:       899.99,
			Stock:       50,
			Category:    "Casa e Cozinha",
			SKU:         "PHILIPS-AIRFRYER-XL-005",
			Active:      true,
			ImageURL:    "https://example.com/airfryer.jpg",
		},
		{
			Name:        "JBL Charge 5",
			Description: "Caixa de som Bluetooth portátil à prova d'água com 20h de bateria.",
			Price:       599.99,
			Stock:       40,
			Category:    "Eletrônicos",
			SKU:         "JBL-CHARGE-5-006",
			Active:      true,
			ImageURL:    "https://example.com/jbl.jpg",
		},
		{
			Name:        "Nike Air Max 90",
			Description: "Tênis Nike Air Max 90 original, conforto e estilo para o dia a dia.",
			Price:       499.99,
			Stock:       60,
			Category:    "Moda e Calçados",
			SKU:         "NIKE-AIRMAX-90-007",
			Active:      true,
			ImageURL:    "https://example.com/nike.jpg",
		},
		{
			Name:        "Kindle Paperwhite",
			Description: "E-reader à prova d'água com tela de 6.8 polegadas e iluminação ajustável.",
			Price:       449.99,
			Stock:       35,
			Category:    "Livros",
			SKU:         "AMAZON-KINDLE-PW-008",
			Active:      true,
			ImageURL:    "https://example.com/kindle.jpg",
		},
	}

	for _, product := range products {
		if err := db.Create(&product).Error; err != nil {
			return fmt.Errorf("failed to seed product %s:%w", product.Name, err)
		}
	}

	log.Printf("Seeded %d products successfully!", len(products))
	return nil
}

func SeedAdminUser(db *gorm.DB) error {
	log.Println("Staring admin user seeding...")

	adminEmail := "aetherraito@protonmail.com"
	adminPassword := "admin_password_123"
	adminUsername := "Administrator"

	var adminUser models.User
	err := db.Where("email = ?", adminEmail).First(&adminUser).Error
	if err == nil {
		log.Println("Admin user already exits, skipping...")
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	admin := models.User{
		Email:    adminEmail,
		Password: string(hashedPassword),
		Name:     adminUsername,
		Role:     "admin",
		Active:   true,
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Printf("Admin %s created", adminUsername)
	log.Printf("Email: %s", adminEmail)
	log.Printf("Password: %s", adminPassword)

	return nil
}
