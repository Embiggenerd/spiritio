package users

import (
	"github.com/Embiggenerd/spiritio/pkg/db"
	"github.com/Embiggenerd/spiritio/pkg/utils"
)

type UsersStore interface {
	CreateUser(admin bool) (*User, error)
	getUserByID(id uint) (*User, error)
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

	newUser := &User{}
	newUser.Name = utils.RandName()
	newUser.Admin = i

	userResult := u.db.DB.Create(newUser)
	return newUser, userResult.Error
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
