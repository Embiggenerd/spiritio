package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/Embiggenerd/spiritio/pkg/config"
	"github.com/Embiggenerd/spiritio/pkg/db"
	"github.com/Embiggenerd/spiritio/pkg/logger"
	"github.com/Embiggenerd/spiritio/pkg/utils"
	jwt "github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

// We only use email and userName
// We only email to authorize
// Names are auto-generated and can be changed
// Adding an email sends a verification email, once verified user gets a green check

// Client sends accessToken unless told it's expired, then sends userToken and gets new accessToken

// All users start out with unverified account
// They are informed this account will be erased in 60 days unless they add an email and verify

// UserToken decrypted has 'secret' phrase, expiration date, and userID that can be lookedup in DB

// How user auth works:
// If client has no userToken or accessToken:
// sends blank, and server sends userToken and accessToken
func NewUsersService(ctx context.Context, cfg *config.Config, log logger.Logger, db *db.Database) Users {
	db.DB.AutoMigrate(&User{})
	// db.DB.AutoMigrate(&UserToken{})
	// db.DB.AutoMigrate(&AccessToken{})

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
	// GetUser(accessToken, userToken string, c client.WebsocketClient) *User
	CreateUser(admin bool) (*User, string, error)
	createJWT(user *User) (string, error)
	GetUserFromAccessToken(token any) (*User, error)
	ValidateAccessToken(tokenString string) (*jwt.Token, error)
	CreateAccessToken(user *User) (string, error)
}

type UsersService struct {
	storage UsersStore
	log     logger.Logger
	cfg     *config.Config
}

func (s *UsersService) CreateUser(admin bool) (*User, string, error) {
	user, err := s.storage.CreateUser(admin)
	if err != nil {
		return nil, "", err
	}
	accessToken, err := s.createJWT(user)
	return user, accessToken, err
}

func (s *UsersService) CreateAccessToken(user *User) (string, error) {
	return s.createJWT(user)
}

func (u *UsersService) createJWT(user *User) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt": 15000,
		"userID":    user.ID,
	}

	secret := u.cfg.AccessTokenSecret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func (s *UsersService) ValidateAccessToken(tokenString string) (*jwt.Token, error) {
	permissionDeniedError := errors.New("permission denied")
	token, err := s.validateJWT(tokenString)
	if err != nil {
		// permissionDenied(w)
		return nil, permissionDeniedError
	}
	if !token.Valid {
		return nil, permissionDeniedError
	}
	return token, err

}

func (s *UsersService) getUserIDFromJWT(token any) (uint, error) {
	t := token.(*jwt.Token)
	claims := t.Claims.(jwt.MapClaims)
	userID := claims["userID"]
	return utils.Float64ToUint(userID.(float64))
}

func (s *UsersService) GetUserFromAccessToken(token any) (*User, error) {
	userID, err := s.getUserIDFromJWT(token)
	if err != nil {
		return nil, err
	}
	return s.storage.getUserByID(userID)
}

func (s *UsersService) getUserByID(id string) (*User, error) {
	userID, err := utils.StringToUint(id)
	if err != nil {
		return nil, err
	}
	return s.storage.getUserByID(userID)
}

func (s *UsersService) validateJWT(tokenString string) (*jwt.Token, error) {
	secret := s.cfg.AccessTokenSecret
	s.log.Info("validating token")
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

// func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) (*jwt.Token, error) {
// 	if r.Method != "POST" {
// 		return fmt.Errorf("method not allowed %s", r.Method)
// 	}

// 	// receve password
// 	// var req LoginRequest
// 	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 	// 	return err
// 	// }

// 	// get user by id
// 	// acc, err := s.store.GetAccountByNumber(int(req.Number))
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	if !acc.ValidPassword(req.Password) {
// 		return fmt.Errorf("not authenticated")
// 	}

// 	token, err := createJWT(acc)
// 	if err != nil {
// 		return err
// 	}

// 	// resp := LoginResponse{
// 	// 	Token:  token,
// 	// 	Number: acc.Number,
// 	// }

// 	// return WriteJSON(w, http.StatusOK, resp)
// }

// func (s *UsersService) CreateUserToken(u *User) (*UserToken, error) {
// 	userToken, err := s.storage.CreateUserToken(u)
// }

// func (u *UsersService) GetUser(accessToken, userToken string, c client.WebsocketClient) *User {
// 	if accessToken == "" {
// 		if userToken == "" {
// 			// create new user
// 		}
// 		// if userToken is expired, send

// 		// refresh access token using userToken
// 	}
// 	return nil
// }

type User struct {
	gorm.Model
	ID              uint `gorm:"primaryKey"`
	Name            string
	Email           string
	Verified        int `gorm:"default:0"`
	Admin           int `gorm:"default:0"`
	UserTokenSecret string
}

// type UserToken struct {
// 	gorm.Model
// 	ExpirationDate time.Time
// 	Secret         string
// 	Nonce          string
// 	Depracated     int `gorm:"default:0"`
// 	AccessToken    string
// 	// UserID         uint
// 	// User           *User `gorm:"foreignKey:UserID"`
// }

// type AccessToken struct {
// 	gorm.Model
// 	ExpirationDate time.Time
// 	Secret         string
// 	Nonce          string
// 	Depracated     int `gorm:"default:0"`
// 	Value          string
// 	// UserID         uint
// 	// User           *User `gorm:"foreignKey:UserID"`
// }

// func encrypt() {
// 	key, _ := hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")
// 	plaintext := []byte("exampleplaintext")

// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
// 	nonce := make([]byte, 12)
// 	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
// 		panic(err.Error())
// 	}

// 	aesgcm, err := cipher.NewGCM(block)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
// 	fmt.Printf("%x\n", ciphertext)
// }

// func decrypt() {
// 	key, _ := hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")
// 	ciphertext, _ := hex.DecodeString("c3aaa29f002ca75870806e44086700f62ce4d43e902b3888e23ceff797a7a471")
// 	nonce, _ := hex.DecodeString("64a9433eae7ccceee2fc0eda")

// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	aesgcm, err := cipher.NewGCM(block)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	fmt.Printf("%s\n", plaintext)
// }
