package handlers

import "time"

// 类型包装器仅用于 Swagger 文档生成，使用显式字段以避免引用外部嵌入类型（例如 gorm.Model）

// NoteSwagger 用于 Swagger 显示 Note 数据结构（简化版）
type NoteSwagger struct {
	ID         uint         `json:"id" example:"1"`
	Title      string       `json:"title" example:"Hello world"`
	Summary    string       `json:"summary" example:"A short summary"`
	Content    string       `json:"content" example:"Detailed content of the note..."`
	CoverImage string       `json:"cover_image" example:"/static/images/cover.webp"`
	AuthorID   uint         `json:"author_id" example:"1"`
	Author     *UserSwagger `json:"author,omitempty"`
	Tags       []TagSwagger `json:"tags,omitempty"`
	IsPublic   bool         `json:"is_public" example:"true"`
	Views      int64        `json:"views" example:"123"`
	Likes      int64        `json:"likes" example:"10"`
	CreatedAt  time.Time    `json:"createdAt"`
	UpdatedAt  time.Time    `json:"updatedAt"`
}

// UserSwagger 用于 Swagger 显示 User（简化版）
type UserSwagger struct {
	ID        uint      `json:"id" example:"1"`
	Username  string    `json:"username" example:"alice"`
	Email     string    `json:"email" example:"user@example.com"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TagSwagger 用于 Swagger 显示 Tag（简化版）
type TagSwagger struct {
	ID        uint      `json:"id" example:"1"`
	Name      string    `json:"name" example:"tech"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ImageMetaSwagger 用于 Swagger 显示 ImageMeta
type ImageMetaSwagger struct {
	URL     string `json:"url" example:"/static/images/abc.webp"`
	Path    string `json:"path"`
	Size    int    `json:"size"`
	ModTime string `json:"mod_time"`
}
