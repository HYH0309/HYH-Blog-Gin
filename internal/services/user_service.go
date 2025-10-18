package services

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"HYH-Blog-Gin/internal/models"
)

// UserService 定义用户相关的业务方法
type UserService interface {
	Register(email, username, password string) (*models.User, error)
	Authenticate(identifier, password string) (*models.User, error)
	GetByID(id uint) (*models.User, error)
}

// 定义常见错误
var (
	ErrUserNotFound = errors.New("user not found")
	ErrBadCreds     = errors.New("invalid credentials")
)

type userService struct {
	users models.UserRepository
}

// NewUserService 创建 UserService 实例
func NewUserService(users models.UserRepository) UserService {
	return &userService{users: users}
}

// Register 注册新用户，密码会被哈希处理
func (s *userService) Register(email, username, password string) (*models.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{Email: email, Username: username, Password: string(hashed)}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}
	// 返回一个去除密码的用户副本
	return user.WithoutPassword(), nil
}

// Authenticate 验证用户凭据（邮箱或用户名 + 密码）
func (s *userService) Authenticate(identifier, password string) (*models.User, error) {
	var user *models.User
	var err error
	if identifier == "" {
		return nil, ErrBadCreds
	}
	if strings.Contains(identifier, "@") {
		user, err = s.users.FindByEmail(identifier)
	} else {
		user, err = s.users.FindByUsername(identifier)
	}
	if err != nil || user == nil || user.ID == 0 {
		return nil, ErrBadCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrBadCreds
	}
	// 返回去除密码的用户副本
	return user.WithoutPassword(), nil
}

// GetByID 根据用户 ID 获取用户信息
func (s *userService) GetByID(id uint) (*models.User, error) {
	user, err := s.users.FindByID(id)
	if err != nil || user == nil || user.ID == 0 {
		return nil, ErrUserNotFound
	}
	// 返回去除密码的用户副本
	return user.WithoutPassword(), nil
}
