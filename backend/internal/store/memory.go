package store

import (
	"slices"

	"home-decision/backend/internal/model"
)

type MemoryStore struct {
	meta    model.Meta
	weights map[string][]model.WeightProfile
	houses  map[string][]model.House
	users   map[string]UserAuthRecord
	tokens  map[string]string
}

func NewMemoryStore(meta model.Meta, weights []model.WeightProfile, houses []model.House) *MemoryStore {
	return &MemoryStore{
		meta: meta,
		weights: map[string][]model.WeightProfile{
			"demo-family": cloneProfiles(weights),
		},
		houses: map[string][]model.House{
			"demo-family": cloneHouses(houses),
		},
		users:  map[string]UserAuthRecord{},
		tokens: map[string]string{},
	}
}

func (m *MemoryStore) GetMeta() model.Meta {
	return m.meta
}

func (m *MemoryStore) GetWeights(householdID string) ([]model.WeightProfile, error) {
	return cloneProfiles(m.weights[householdID]), nil
}

func (m *MemoryStore) SaveWeights(householdID string, profiles []model.WeightProfile) error {
	m.weights[householdID] = cloneProfiles(profiles)
	return nil
}

func (m *MemoryStore) ListHouses(householdID string) ([]model.House, error) {
	return cloneHouses(m.houses[householdID]), nil
}

func (m *MemoryStore) GetHouse(householdID, houseID string) (*model.House, error) {
	for _, house := range m.houses[householdID] {
		if house.ID == houseID {
			copyHouse := house
			return &copyHouse, nil
		}
	}
	return nil, ErrHouseNotFound
}

func (m *MemoryStore) CreateHouse(householdID string, house model.House) (*model.House, error) {
	m.houses[householdID] = append([]model.House{house}, m.houses[householdID]...)
	copyHouse := house
	return &copyHouse, nil
}

func (m *MemoryStore) UpdateHouse(householdID, houseID string, house model.House) (*model.House, error) {
	for index, item := range m.houses[householdID] {
		if item.ID == houseID {
			m.houses[householdID][index] = house
			copyHouse := house
			return &copyHouse, nil
		}
	}
	return nil, ErrHouseNotFound
}

func (m *MemoryStore) DeleteHouse(householdID, houseID string) error {
	items := m.houses[householdID]
	index := slices.IndexFunc(items, func(item model.House) bool {
		return item.ID == houseID
	})
	if index < 0 {
		return ErrHouseNotFound
	}
	m.houses[householdID] = append(items[:index], items[index+1:]...)
	return nil
}

func (m *MemoryStore) CreateUser(user model.User, passwordSalt, passwordHash, householdID string) error {
	if len(m.users) == 0 {
		user.IsAdmin = true
	}
	m.users[user.ID] = UserAuthRecord{
		User:         user,
		PasswordSalt: passwordSalt,
		PasswordHash: passwordHash,
	}
	if _, ok := m.weights[householdID]; !ok {
		m.weights[householdID] = defaultProfiles()
	}
	if _, ok := m.houses[householdID]; !ok {
		m.houses[householdID] = []model.House{}
	}
	return nil
}

func (m *MemoryStore) FindUserAuthByLoginID(loginID string) (*UserAuthRecord, error) {
	for _, record := range m.users {
		if record.User.LoginID == loginID {
			copyRecord := record
			return &copyRecord, nil
		}
	}
	return nil, nil
}

func (m *MemoryStore) FindUserByLoginID(loginID string) (*model.User, error) {
	record, err := m.FindUserAuthByLoginID(loginID)
	if err != nil || record == nil {
		return nil, err
	}
	user := record.User
	return &user, nil
}

func (m *MemoryStore) FindUserByID(userID string) (*model.User, error) {
	record, ok := m.users[userID]
	if !ok {
		return nil, nil
	}
	user := record.User
	return &user, nil
}

func (m *MemoryStore) FindUserByLinkCode(linkCode string) (*model.User, error) {
	for _, record := range m.users {
		if record.User.LinkCode == linkCode {
			user := record.User
			return &user, nil
		}
	}
	return nil, nil
}

func (m *MemoryStore) CreateSession(session model.Session) error {
	m.tokens[session.Token] = session.UserID
	return nil
}

func (m *MemoryStore) FindUserBySessionToken(token string) (*model.User, error) {
	userID, ok := m.tokens[token]
	if !ok {
		return nil, nil
	}
	return m.FindUserByID(userID)
}

func (m *MemoryStore) DeleteSession(token string) error {
	delete(m.tokens, token)
	return nil
}

func (m *MemoryStore) FindHouseholdIDByUserID(userID string) (string, error) {
	return "demo-family", nil
}

func (m *MemoryStore) ListAccountMembers(householdID string) ([]model.AccountMember, error) {
	var members []model.AccountMember
	for _, record := range m.users {
		members = append(members, model.AccountMember{
			UserID:      record.User.ID,
			LoginID:     record.User.LoginID,
			DisplayName: record.User.DisplayName,
			LinkCode:    record.User.LinkCode,
		})
	}
	return members, nil
}

func (m *MemoryStore) LinkUsers(userID, partnerUserID string) error {
	return nil
}

func (m *MemoryStore) ListAdminUsers() ([]model.AdminUser, error) {
	var items []model.AdminUser
	for _, record := range m.users {
		items = append(items, model.AdminUser{
			User:        record.User,
			HouseholdID: "demo-family",
		})
	}
	return items, nil
}

func (m *MemoryStore) SetUserAdmin(userID string, isAdmin bool) error {
	record, ok := m.users[userID]
	if !ok {
		return nil
	}
	record.User.IsAdmin = isAdmin
	m.users[userID] = record
	return nil
}

func cloneProfiles(in []model.WeightProfile) []model.WeightProfile {
	out := make([]model.WeightProfile, 0, len(in))
	for _, profile := range in {
		weights := make(map[string]float64, len(profile.Weights))
		for key, value := range profile.Weights {
			weights[key] = value
		}
		out = append(out, model.WeightProfile{
			Role:    profile.Role,
			Label:   profile.Label,
			Weights: weights,
		})
	}
	return out
}

func cloneHouses(in []model.House) []model.House {
	out := make([]model.House, 0, len(in))
	for _, house := range in {
		copyHouse := house
		copyHouse.BonusSelections = slices.Clone(house.BonusSelections)
		copyHouse.RiskSelections = slices.Clone(house.RiskSelections)
		out = append(out, copyHouse)
	}
	return out
}
