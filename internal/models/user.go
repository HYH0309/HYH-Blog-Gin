package models

import "gorm.io/gorm"

// User 用户模型
type User struct {
	gorm.Model
	Email    string `json:"email" gorm:"uniqueIndex;not null"`
	Username string `json:"username" gorm:"uniqueIndex;not null"`
	Password string `json:"-" gorm:"not null"`
	Notes    []Note `json:"notes,omitempty" gorm:"foreignKey:AuthorID"`
}

func (User) TableName() string { return "users" }

// WithoutPassword 返回一个不包含密码的用户副本
func (u User) WithoutPassword() *User {
	u.Password = ""
	return &u
}

// UserRepository 用户数据操作接口
type UserRepository interface {
	Create(user *User) error
	FindByID(id uint) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id uint) error
}
