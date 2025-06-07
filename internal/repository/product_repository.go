package repository

import (
	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (r *ProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepository) GetAll() ([]models.Product, int, error) {
	var products []models.Product
	err := r.db.Where("active = ?", true).Find(&products).Error
	return products, len(products), err
}

func (r *ProductRepository) GetWithFilters(page, limit int, category, search string) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{}).Where("active = ?", true)

	if category != "" {
		query = query.Where("category ILIKE ?", "%"+category+"%")
	}

	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&products).Error

	return products, total, err
}

func (r *ProductRepository) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.Where("id = ? AND active = ?", id, true).First(&product, id).Error
	return &product, err
}

func (r *ProductRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}
func (r *ProductRepository) Delete(id uint) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).Update("active", false).Error
}

func (r *ProductRepository) HardDelete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}

func (r *ProductRepository) GetByCategory(category string) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("category ILIKE ? AND active = ?", "%"+category+"%", true).Find(&products).Error
	return products, err
}

func (r *ProductRepository) SearchByName(name string) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("name <= ? AND active = ?", "%"+name+"%", true).Find(&products).Error
	return products, err
}

func (r *ProductRepository) GetLowStock(threshold int) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("stock <= ? AND active = ?", threshold, true).Find(&products).Error
	return products, err
}

func (r *ProductRepository) GetByPriceRange(minPrice, maxPrice float64) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("price BETWEEN ? AND ? AND active = ?", minPrice, maxPrice, true).Find(&products).Error
	return products, err
}

func (r *ProductRepository) GetMostExpensive(limit int) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("active = ?", true).Order("price DESC").Limit(limit).Find(&products).Error
	return products, err
}

func (r *ProductRepository) GetNewest(limit int) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("active = ?", true).Order("created_at DESC").Limit(limit).Find(&products).Error
	return products, err
}

func (r *ProductRepository) GetCategories() ([]string, error) {
	var categories []string
	err := r.db.Model(&models.Product{}).Where("active = ?", true).Distinct("category").Pluck("category", &categories).Error
	return categories, err
}

func (r *ProductRepository) CountByCategory() (map[string]int64, error) {
	type CategoryCount struct {
		Category string
		Count    int64
	}

	var results []CategoryCount
	err := r.db.Model(&models.Product{}).
		Select("category, count(*) as count").
		Where("active = ?", true).
		Group("category").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	categoryMap := make(map[string]int64)
	for _, result := range results {
		categoryMap[result.Category] = result.Count
	}

	return categoryMap, nil
}

func (r *ProductRepository) UpdateStock(id uint, stock int) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).Update("stock", stock).Error
}

func (r *ProductRepository) IncrementStock(id uint, quantity int) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).Update("stock", gorm.Expr("stock + ?", quantity)).Error
}

func (r *ProductRepository) DecrementStock(id uint, quantity int) error {
	return r.db.Model(&models.Product{}).Where("id = ? AND stock >= ?", id, quantity).Update("stock", gorm.Expr("stock - ?", quantity)).Error
}
