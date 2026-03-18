package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"home-decision/backend/internal/model"
	"home-decision/backend/internal/store"
)

var (
	ErrInvalidCredentials = errors.New("login id 或密码错误")
	ErrLoginExists        = errors.New("该账号已存在")
	ErrInvalidLinkCode    = errors.New("关联 ID 不存在")
	ErrLinkSelf           = errors.New("不能关联自己")
	ErrUnauthorized       = errors.New("unauthorized")
)

type AuthService struct {
	store store.Store
}

func NewAuthService(s store.Store) *AuthService {
	return &AuthService{store: s}
}

func (s *AuthService) Register(loginID, password, displayName string) (*model.User, string, error) {
	existing, err := s.store.FindUserByLoginID(loginID)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", ErrLoginExists
	}

	user := model.User{
		ID:          newID("user"),
		LoginID:     loginID,
		DisplayName: displayName,
		// 关联码是夫妻双方线下互相输入的唯一 ID，不直接暴露数据库主键。
		LinkCode: upperHex(4),
	}
	salt := randomHex(16)
	hash := hashPassword(password, salt)
	householdID := newID("household")

	if err := s.store.CreateUser(user, salt, hash, householdID); err != nil {
		return nil, "", err
	}
	token, err := s.createSession(user.ID)
	if err != nil {
		return nil, "", err
	}
	return &user, token, nil
}

func (s *AuthService) Login(loginID, password string) (*model.User, string, error) {
	record, err := s.store.FindUserAuthByLoginID(loginID)
	if err != nil {
		return nil, "", err
	}
	if record == nil || record.PasswordHash != hashPassword(password, record.PasswordSalt) {
		return nil, "", ErrInvalidCredentials
	}
	token, err := s.createSession(record.User.ID)
	if err != nil {
		return nil, "", err
	}
	return &record.User, token, nil
}

func (s *AuthService) Logout(token string) error {
	return s.store.DeleteSession(token)
}

func (s *AuthService) ProfileByToken(token string) (*model.AccountProfile, error) {
	user, err := s.store.FindUserBySessionToken(token)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUnauthorized
	}
	householdID, err := s.store.FindHouseholdIDByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	members, err := s.store.ListAccountMembers(householdID)
	if err != nil {
		return nil, err
	}
	return &model.AccountProfile{
		User:        *user,
		HouseholdID: householdID,
		Members:     members,
	}, nil
}

func (s *AuthService) LinkPartner(token, partnerLinkCode string) (*model.AccountProfile, error) {
	user, err := s.store.FindUserBySessionToken(token)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUnauthorized
	}
	partner, err := s.store.FindUserByLinkCode(partnerLinkCode)
	if err != nil {
		return nil, err
	}
	if partner == nil {
		return nil, ErrInvalidLinkCode
	}
	if partner.ID == user.ID {
		return nil, ErrLinkSelf
	}
	// 关联成功后，两个人会并入同一个 household，后续房源和权重都会共享。
	if err := s.store.LinkUsers(user.ID, partner.ID); err != nil {
		return nil, err
	}
	return s.ProfileByToken(token)
}

func (s *AuthService) AdminUsers(token string) ([]model.AdminUser, error) {
	user, err := s.store.FindUserBySessionToken(token)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUnauthorized
	}
	if !user.IsAdmin {
		return nil, errors.New("forbidden")
	}
	return s.store.ListAdminUsers()
}

func (s *AuthService) SetAdmin(token, targetUserID string, isAdmin bool) error {
	user, err := s.store.FindUserBySessionToken(token)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUnauthorized
	}
	if !user.IsAdmin {
		return errors.New("forbidden")
	}
	if user.ID == targetUserID && !isAdmin {
		// 防止唯一管理员把自己也降权，导致后台失去管理入口。
		return errors.New("不能取消自己的管理员权限")
	}
	return s.store.SetUserAdmin(targetUserID, isAdmin)
}

func (s *AuthService) createSession(userID string) (string, error) {
	// 当前实现使用服务端 session token，前端只存 token，不直接持有用户敏感信息。
	token := randomHex(24)
	expiresAt := time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02 15:04:05")
	if err := s.store.CreateSession(model.Session{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}); err != nil {
		return "", err
	}
	return token, nil
}

func hashPassword(password, salt string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", salt, password)))
	return hex.EncodeToString(sum[:])
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, randomHex(8))
}

func randomHex(size int) string {
	buf := make([]byte, size)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func upperHex(size int) string {
	raw := randomHex(size)
	out := make([]byte, len(raw))
	for i := range raw {
		c := raw[i]
		if c >= 'a' && c <= 'f' {
			out[i] = c - 32
		} else {
			out[i] = c
		}
	}
	return string(out)
}
