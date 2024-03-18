package users

import (
	"fmt"

	"github.com/Embiggenerd/spiritio/pkg/db"
)

type UsersStore interface {
	CreateUser(admin bool) (*User, error)
	getUserByID(id uint) (*User, error)
	UpdateUserName(id uint, name string) error
	getUserByName(name string) (*User, error)
	updateUserPassword(id uint, password string) error
	// CreateUserToken(u *User) (*UserToken, error)
	// SaveChatlog(text string, room *Room) *ChatLog
	// FindRByID(roomID uint) (*User, error)
}

type UsersStorage struct {
	db *db.Database
}

func (s *UsersStorage) getUserByID(userID uint) (*User, error) {
	foundUser := &User{}
	userResult := s.db.DB.Where(User{ID: userID}).First(foundUser)
	return foundUser, userResult.Error
}

func (u *UsersStorage) CreateUser(admin bool) (*User, error) {
	i := 0
	if admin {
		i = 1
	}

	newUser := &User{
		Admin: i,
	}

	userResult := u.db.DB.Create(newUser)
	return newUser, userResult.Error
}

func (u *UsersStorage) UpdateUserName(id uint, name string) error {
	result := u.db.DB.Model(&User{ID: id}).Updates(User{Name: name})
	fmt.Println("updateUserName*", id, name, result.Error)
	return result.Error
}

func (u *UsersStorage) getUserByName(name string) (*User, error) {
	foundUser := &User{}
	userResult := u.db.DB.Where(User{Name: name}).First(foundUser)
	return foundUser, userResult.Error
}

func (u *UsersStorage) updateUserPassword(id uint, password string) error {
	result := u.db.DB.Where(User{ID: id}).Updates(User{Password: password})
	fmt.Println("%$%$%", id, password)
	return result.Error
}

// func (s *UsersStorage) CreateUserToken(u *User) (*UserToken, error) {
// 	// make uuid
// 	uuid := uuid.New()
// 	// foundUser := &User{}
// 	roomResult := s.db.DB.Where(User{UserTokenSecret: uuid.String()}) // check for uniqueness
// 	roomResult.Count()
// 	// set userToken 'value' on user
// 	// encrypt uuid as 'value' of userToken
// 	ut := &UserToken{}
// }

// func ensureUnique(uuid string) {
// 	result := s.db.DB.Where(User{UserToken: uuid}) // check for uniqueness
// 	roomResult.Count()
// }
