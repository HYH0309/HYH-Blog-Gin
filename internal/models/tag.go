package models

import "gorm.io/gorm"

// Tag 标签模型
type Tag struct {
	gorm.Model
	Name  string `json:"name" gorm:"uniqueIndex;not null"`
	Notes []Note `json:"notes,omitempty" gorm:"many2many:note_tags;"`
}

func (Tag) TableName() string { return "tags" }

// TagRepository 标签数据操作接口
type TagRepository interface {
	Create(tag *Tag) error
	FindByID(id uint) (*Tag, error)
	FindByName(name string) (*Tag, error)
	FindOrCreate(names []string) ([]Tag, error)
	FindByNote(noteID uint) ([]Tag, error)
	Update(tag *Tag) error
	Delete(id uint) error
}
