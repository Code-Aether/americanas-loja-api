package services

import (
	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"github.com/Code-Aether/americanas-loja-api/internal/repository"
	"github.com/go-redis/redis/v8"
)

type ProductService struct {
	productRepo *repository.ProductRepository
	redis       *redis.Client
}

func NewProductService(productRepo *repository.ProductRepository, redis *redis.Client) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		redis:       redis,
	}
}

func (s *ProductService) GetAll() ([]models.Product, error) {
	return s.productRepo.GetAll()
}

func (s *ProductService) GetByID(id uint) (*models.Product, error) {
	return s.productRepo.GetByID(id)
}

func (s *ProductService) Create(product *models.Product) error {
	return s.productRepo.Create(product)
}

func (s *ProductService) Update(product *models.Product) error {
	return s.productRepo.Update(product)
}

func (s *ProductService) Delete(id uint) error {
	return s.productRepo.Delete(id)
}
