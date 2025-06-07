package repository

import (
	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Find(&users).Error
	return users, err
}

func (r *UserRepository) GetAllWithPagination(page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := r.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error

	return users, total, err
}

func (r *UserRepository) GetActiveUsers() ([]models.User, error) {
	var users []models.User
	err := r.db.Where("active = ?", true).Find(&users).Error
	return users, err
}

func (r *UserRepository) GetByRole(role string) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("role = ?", role).Find(&users).Error
	return users, err
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *UserRepository) Deactivate(id uint) error {
	return r.getUserById(id).Update("activate", false).Error
}

func (r *UserRepository) Activate(id uint) error {
	return r.getUserById(id).Update("active = ?", true).Error
}

func (r *UserRepository) UpdatePassword(id uint, hashedPassword string) error {
	return r.getUserById(id).Update("password", hashedPassword).Error
}

func (r *UserRepository) UpdateRole(id uint, role string) error {
	return r.getUserById(id).Update("role", role).Error
}

func (r *UserRepository) UpdateLastLogin(id uint) error {
	return r.getUserById(id).Update("updated_at", "NOW()").Error
}

func (r *UserRepository) CountUsers() (int64, error) {
	var count int64
	err := r.bindUserModel().Count(&count).Error
	return count, err
}

func (r *UserRepository) CountActiveUsers() (int64, error) {
	var count int64
	err := r.bindUserModel().Where("active = ?", true).Count(&count).Error
	return count, err
}

func (r *UserRepository) SearchUsers(query string) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("name ILIKE ? or email ILIKE ?", "%"+query+"%", "%"+query+"%").Find(&users).Error
	return users, err
}

// Private functions
func (r *UserRepository) bindUserModel() *gorm.DB {
	return r.db.Model(&models.User{})
}

func (r *UserRepository) getUserById(id uint) *gorm.DB {
	return r.bindUserModel().Where("id = ?", id)
}
