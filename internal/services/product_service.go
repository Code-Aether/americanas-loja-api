package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

func (s *ProductService) GetAll(page, limit int, category, search string) ([]models.Product, int64, error) {
	cacheKey := fmt.Sprintf("product:page:%d:limit:%d:category:%s:search:%s",
		page, limit, category, search)

	if s.redis != nil {
		cached, err := s.redis.Get(context.Background(), cacheKey).Result()
		if err == nil {
			var result struct {
				Products []models.Product `json:"products"`
				Total    int64            `json:"total"`
			}
			if json.Unmarshal([]byte(cached), &result) == nil {
				return result.Products, result.Total, nil
			}
		}
	}

	products, total, err := s.productRepo.GetWithFilters(page, limit, category, search)

	if err != nil {
		return nil, 0, err
	}

	if s.redis != nil {
		result := struct {
			Products []models.Product `json:"products"`
			Total    int64            `json:"total"`
		}{
			Products: products,
			Total:    total,
		}

		data, _ := json.Marshal(result)
		s.redis.Set(context.Background(), cacheKey, data, 5*time.Minute)
	}

	return products, total, nil
}

func (s *ProductService) GetByID(id uint) (*models.Product, error) {
	cacheKey := fmt.Sprintf("product:%d", id)

	if s.redis != nil {
		cached, err := s.redis.Get(context.Background(), cacheKey).Result()
		if err == nil {
			var product models.Product
			if json.Unmarshal([]byte(cached), &product) == nil {
				return &product, nil
			}
		}
	}

	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if s.redis != nil {
		data, _ := json.Marshal(product)
		s.redis.Set(context.Background(), cacheKey, data, 10*time.Minute)
	}

	return product, nil
}

func (s *ProductService) Create(product *models.Product) error {
	err := s.productRepo.Create(product)
	if err != nil {
		return err
	}

	s.invalidateListCache()

	return nil
}

func (s *ProductService) Update(product *models.Product) error {
	err := s.productRepo.Update(product)
	if err != nil {
		return err
	}

	s.invalidateProductCache(product.ID)
	s.invalidateListCache()

	return nil
}

func (s *ProductService) Delete(id uint) error {
	err := s.productRepo.Delete(id)
	if err != nil {
		return err
	}

	s.invalidateProductCache(id)
	s.invalidateListCache()

	return nil
}

func (s *ProductService) GetByCategory(category string) ([]models.Product, error) {
	return s.productRepo.GetByCategory(category)
}

func (s *ProductService) SearchProducts(query string) ([]models.Product, error) {
	return s.productRepo.SearchByName(query)
}

func (s *ProductService) UpdateStock(id uint, quantity int) error {
	product, err := s.GetByID(id)
	if err != nil {
		return err
	}

	newStock := product.Stock + quantity
	if newStock < 0 {
		return fmt.Errorf("NOT_ENOUGH_STOCK")
	}

	product.Stock = newStock
	return s.Update(product)
}

func (s *ProductService) invalidateProductCache(id uint) {
	if s.redis != nil {
		cacheKey := fmt.Sprintf("product:%d", id)
		s.redis.Del(context.Background(), cacheKey)
	}
}

func (s *ProductService) invalidateListCache() {
	if s.redis != nil {
		keys, err := s.redis.Keys(context.Background(), "products:*").Result()
		if err == nil && len(keys) > 0 {
			s.redis.Del(context.Background(), keys...)
		}
	}
}
