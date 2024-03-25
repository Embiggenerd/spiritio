package users

import (
	"errors"

	"github.com/Embiggenerd/spiritio/pkg/db"
	"github.com/Embiggenerd/spiritio/pkg/utils"
	"gorm.io/gorm"
)

type UsersStore interface {
	CreateUser(admin bool) (*User, error)
	getUserByID(id uint) (*User, error)
	UpdateUserName(id uint, name string) error
	getUserByName(name string) (*User, error)
	updateUserPassword(id uint, password string) error
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

	name := u.EnsureUniqueName(utils.RandName(), 0)

	newUser := &User{
		Admin: i,
		Name:  name,
	}

	userResult := u.db.DB.Create(newUser)
	return newUser, userResult.Error
}

func (u *UsersStorage) UpdateUserName(id uint, name string) error {
	result := u.db.DB.Model(&User{ID: id}).Updates(User{Name: name})
	return result.Error
}

func (u *UsersStorage) EnsureUniqueName(name string, id uint) string {
	user, err := u.getUserByName(name)
	if errors.Is(err, gorm.ErrRecordNotFound) || user.ID == id {
		return name
	}

	return u.EnsureUniqueName(utils.RandName(), id)
}

func (u *UsersStorage) getUserByName(name string) (*User, error) {
	foundUser := &User{}
	userResult := u.db.DB.Where(User{Name: name}).First(foundUser)
	return foundUser, userResult.Error
}

func (u *UsersStorage) updateUserPassword(id uint, password string) error {
	result := u.db.DB.Where(User{ID: id}).Updates(User{Password: password})
	return result.Error
}
