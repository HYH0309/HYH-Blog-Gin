package models

import (
	"gorm.io/gorm"
)

// Note 笔记模型
type Note struct {
	gorm.Model
	Title      string `json:"title" gorm:"not null" example:"Hello world"`
	Summary    string `json:"summary" gorm:"type:text" example:"A short summary"`
	Content    string `json:"content" gorm:"type:text;not null" example:"Detailed content of the note..."`
	CoverImage string `json:"cover_image" example:"/static/images/cover.webp"`
	AuthorID   uint   `json:"author_id" gorm:"index" example:"1"`
	Author     User   `json:"author" gorm:"foreignKey:AuthorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Tags       []Tag  `json:"tags" gorm:"many2many:note_tags;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	IsPublic   bool   `json:"is_public" gorm:"default:false;index" example:"true"`
	Views      int64  `json:"views" gorm:"default:0" example:"123"`
	Likes      int64  `json:"likes" gorm:"default:0" example:"10"`
}

func (Note) TableName() string { return "notes" }

// NoteRepository 笔记数据操作接口
type NoteRepository interface {
	Create(note *Note) error
	CreateWithTags(note *Note, tagNames []string) error
	FindByID(id uint) (*Note, error)
	FindByAuthor(authorID uint, page, limit int) ([]Note, int64, error)
	Search(authorID uint, query string, tags []string) ([]Note, error)
	Update(note *Note) error
	UpdateWithTags(note *Note, tagNames []string) error
	Delete(id uint) error
	AddTags(noteID uint, tags []Tag) error
	RemoveTags(noteID uint, tagIDs []uint) error
}
