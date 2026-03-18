package store

import "home-decision/backend/internal/model"

type Store interface {
	GetMeta() model.Meta
	GetWeights(householdID string) ([]model.WeightProfile, error)
	SaveWeights(householdID string, profiles []model.WeightProfile) error
	ListHouses(householdID string) ([]model.House, error)
	GetHouse(householdID, houseID string) (*model.House, error)
	CreateHouse(householdID string, house model.House) (*model.House, error)
	UpdateHouse(householdID, houseID string, house model.House) (*model.House, error)
	DeleteHouse(householdID, houseID string) error
	CreateUser(user model.User, passwordSalt, passwordHash, householdID string) error
	FindUserAuthByLoginID(loginID string) (*UserAuthRecord, error)
	FindUserByLoginID(loginID string) (*model.User, error)
	FindUserByID(userID string) (*model.User, error)
	FindUserByLinkCode(linkCode string) (*model.User, error)
	CreateSession(session model.Session) error
	FindUserBySessionToken(token string) (*model.User, error)
	DeleteSession(token string) error
	FindHouseholdIDByUserID(userID string) (string, error)
	ListAccountMembers(householdID string) ([]model.AccountMember, error)
	LinkUsers(userID, partnerUserID string) error
	ListAdminUsers() ([]model.AdminUser, error)
	SetUserAdmin(userID string, isAdmin bool) error
}

type UserAuthRecord struct {
	User         model.User
	PasswordSalt string
	PasswordHash string
}
