package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Embiggenerd/spiritio/pkg/config"
	"github.com/Embiggenerd/spiritio/pkg/db"
	"github.com/Embiggenerd/spiritio/pkg/logger"
	"github.com/Embiggenerd/spiritio/pkg/utils"
	jwt "github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func NewUsersService(ctx context.Context, cfg *config.Config, log logger.Logger, db *db.Database) Users {
	db.DB.AutoMigrate(&User{})

	usersStorage := &UsersStorage{
		db: db,
	}

	usersService := &UsersService{
		storage: usersStorage,
		log:     log,
		cfg:     cfg,
	}

	return usersService
}

type Users interface {
	CreateUser(admin bool) (*User, string, error)
	GetUserFromAccessToken(token any) (*User, error)
	ValidateAccessToken(tokenString string) (*jwt.Token, error)
	CreateAccessToken(user *User) (string, error)
	UpdateUserName(id uint, name string) error
	UpdateUserPassword(id uint, password string) error
	EnsureUnique(name string, id uint) string
	GetUserByName(name string) (*User, error)
	ValidateNamePassword(name, password string) (*User, error)
	GetUserByID(id uint) (*User, error)
}

type UsersService struct {
	storage UsersStore
	log     logger.Logger
	cfg     *config.Config
}

type CustomClaims struct {
	UserID uint
	jwt.RegisteredClaims
}

func (s *UsersService) CreateUser(admin bool) (*User, string, error) {
	user, err := s.storage.CreateUser(admin)
	if err != nil {
		return nil, "", err
	}
	accessToken, err := s.createJWT(user)
	return user, accessToken, err
}

func (s *UsersService) UpdateUserName(id uint, name string) error {
	if name == "" {
		return fmt.Errorf("name is not valid")
	}
	_, err := s.GetUserByName(name)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("name is not unique")
	}
	return s.storage.UpdateUserName(id, name)
}

func (s *UsersService) CreateAccessToken(user *User) (string, error) {
	return s.createJWT(user)
}

func (u *UsersService) createJWT(user *User) (string, error) {
	claims := CustomClaims{
		user.ID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	secret := u.cfg.AccessTokenSecret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func (s *UsersService) ValidateAccessToken(tokenString string) (*jwt.Token, error) {
	token, err := s.validateJWT(tokenString)
	if err != nil {
		return nil, err

	}

	claims := token.Claims.(*CustomClaims)
	fmt.Println("claimzz", claims.ExpiresAt)
	expired := claims.ExpiresAt.Time.Before(time.Now())

	if !token.Valid || expired {
		err = errors.New("permission denied")
	}

	return token, err
}

func (s *UsersService) getUserIDFromJWT(token any) uint {
	t := token.(*jwt.Token)
	claims := t.Claims.(*CustomClaims)
	userID := claims.UserID
	return userID
}

func (s *UsersService) GetUserFromAccessToken(token any) (*User, error) {
	userID := s.getUserIDFromJWT(token)
	return s.storage.getUserByID(userID)
}

func (s *UsersService) GetUserByID(id uint) (*User, error) {
	return s.storage.getUserByID(id)
}

func (s *UsersService) GetUserByName(name string) (*User, error) {
	return s.storage.getUserByName(name)
}

// EnsureUnique checks if we have a globally unique name
func (u *UsersService) EnsureUnique(name string, id uint) string {
	user, err := u.GetUserByName(name)
	if errors.Is(err, gorm.ErrRecordNotFound) || user.ID == id {
		return name
	}

	return u.EnsureUnique(utils.RandName(), id)
}

func (u *UsersService) UpdateUserPassword(id uint, password string) error {
	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return u.storage.updateUserPassword(id, string(encpw))
}

func (s *UsersService) validateJWT(tokenString string) (*jwt.Token, error) {
	secret := s.cfg.AccessTokenSecret
	s.log.Info("validating token")
	return jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

func (s *UsersService) ValidateNamePassword(name, password string) (*User, error) {
	user, err := s.GetUserByName(name)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return user, err
}

type User struct {
	gorm.Model
	ID       uint `gorm:"primaryKey"`
	Name     string
	Email    string
	Verified int `gorm:"default:0"`
	Admin    int `gorm:"default:0"`
	Password string
}
