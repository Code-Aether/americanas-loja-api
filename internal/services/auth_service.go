package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Code-Aether/americanas-loja-api/internal/models"
	"github.com/Code-Aether/americanas-loja-api/internal/repository"
	"github.com/Code-Aether/americanas-loja-api/internal/types"
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

func (s *AuthService) Login(req types.LoginRequest) (*string, *models.User, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("invalid credentials")
		}
		return nil, nil, errors.New("internal error")
	}

	if !user.Active {
		return nil, nil, errors.New("user is inactive")
	}

	if !s.checkPassword(req.Password, user.Password) {
		return nil, nil, errors.New("invalid credentials")
	}

	token, _, err := s.GenerateJWT(user)
	if err != nil {
		return nil, nil, errors.New("error generating access token")
	}

	return &token, user, nil
}

func (s *AuthService) Register(user *models.User) (*string, error) {
	if user.Email == "" {
		return nil, errors.New("email is required")
	}

	if len(user.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters long")
	}

	existingUser, err := s.userRepo.GetByEmail(user.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := s.hashPassword(user.Password)
	if err != nil {
		return nil, errors.New("error processing the user password")
	}

	user.Password = hashedPassword

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("error create a new user")
	}

	token, _, err := s.GenerateJWT(user)
	if err != nil {
		return nil, errors.New("error created a new token")
	}

	return &token, nil
}

func (s *AuthService) GenerateJWT(user *models.User) (string, *models.User, error) {
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

	signedString, err := token.SignedString([]byte(s.jwtSecret))
	return signedString, user, err
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

func (s *AuthService) RefreshToken(tokenString string) (string, *models.User, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", nil, err
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return "", nil, errors.New("USER_NOT_FOUND")
	}

	return s.GenerateJWT(user)
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
