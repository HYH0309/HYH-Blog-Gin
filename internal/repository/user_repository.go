package repository

import (
	"HYH-Blog-Gin/internal/models"

	"gorm.io/gorm"
)

// userRepository 提供 UserRepository 接口的 GORM 实现。
// 该实现是无状态的，持有 *gorm.DB 引用，可在多 goroutine 中安全复用。
// 约定：
// - 查询未命中时返回 gorm.ErrRecordNotFound；
// - Create/Update/Delete 出错时直接返回底层 DB 错误；
// - 未进行字段白名单过滤，调用方需保证写入安全。
type userRepository struct{ db *gorm.DB }

// NewUserRepository 构造一个基于 GORM 的用户仓储实现。
func NewUserRepository(db *gorm.DB) models.UserRepository { return &userRepository{db: db} }

// Create 新建用户记录。
// 若存在唯一键冲突（email/username），GORM 将返回错误。
func (r *userRepository) Create(user *models.User) error { return r.db.Create(user).Error }

// FindByID 根据主键查询用户。
// 返回的 user 为零值时，err 通常为 gorm.ErrRecordNotFound。
func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	return &user, err
}

// FindByEmail 根据邮箱查询用户。
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "email = ?", email).Error
	return &user, err
}

// FindByUsername 根据用户名查询用户。
func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "username = ?", username).Error
	return &user, err
}

// Update 保存用户全部字段（根据主键）。
func (r *userRepository) Update(user *models.User) error { return r.db.Save(user).Error }

// Delete 根据主键删除用户。
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}
