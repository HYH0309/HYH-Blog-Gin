package models

import "gorm.io/gorm"

// Tag 标签模型
type Tag struct {
	gorm.Model
	Name  string `json:"name" gorm:"uniqueIndex;not null" example:"tech"`
	Notes []Note `json:"notes,omitempty" gorm:"many2many:note_tags;"`
}

func (Tag) TableName() string { return "tags" }

// TagRepository 标签数据操作接口
type TagRepository interface {
	Create(tag *Tag) error
	FindByID(id uint) (*Tag, error)
	FindByName(name string) (*Tag, error)
	// FindOrCreate 接受一组名称，返回对应的 Tag 列表以及每个项是否为新创建（true = created）
	FindOrCreate(names []string) ([]Tag, []bool, error)
	FindByNote(noteID uint) ([]Tag, error)
	Update(tag *Tag) error
	Delete(id uint) error
	List(page, perPage int) ([]Tag, int64, error)
}
