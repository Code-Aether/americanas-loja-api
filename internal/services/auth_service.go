package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"github.com/Code-Aether/americanas-loja-api/internal/repository"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Login(req models.LoginRequest) (*models.LoginResponse, error) {

	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("EMAIL_OR_PASS_NOT_VALID")
		}
		return nil, errors.New("INTERNAL_ERROR")
	}

	if !user.Active {
		return nil, errors.New("INACTIVE_USER")
	}

	if !s.checkPassword(req.Password, user.Password) {
		return nil, errors.New("EMAIL_OR_PASS_NOT_VALID")
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, errors.New("ERROR_GENERATING_ACCESS_TOKEN")
	}

	user.Password = ""

	return &models.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *AuthService) Register(req models.RegisterRequest) (*models.User, error) {
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user already exsists with this email")
	}

	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, errors.New("error processing the user password")
	}

	user := &models.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
		Role:     "user",
		Active:   true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("error create a new user")
	}

	return user, nil
}

func (s *AuthService) generateJWT(user *models.User) (string, error) {
	timeNow := time.Now()
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(timeNow.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(timeNow),
			NotBefore: jwt.NewNumericDate(timeNow),
			Issuer:    "americanas-loja-api",
			Subject:   user.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("SIGN_METHOD_INVALID")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("INVALID_TOKEN")
}

func (s *AuthService) GetUserByToken(tokenString string) (*models.User, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, errors.New("USER_NOT_FOUND")
	}

	if !user.Active {
		return nil, errors.New("INACTIVE_USER")
	}

	user.Password = ""
	return user, nil
}

func (s *AuthService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return "", errors.New("USER_NOT_FOUND")
	}

	return s.generateJWT(user)
}

func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("USER_NOT_FOUND")
	}

	if !s.checkPassword(oldPassword, user.Password) {
		return errors.New("INCORRECT_PASSWORD")
	}

	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return errors.New("ERROR_WHEN_PROCESSING_CHANGE_PASS")
	}

	user.Password = hashedPassword
	return s.userRepo.Update(user)
}

func (s *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *AuthService) checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
