package handlers

import "HYH-Blog-Gin/internal/models"

// 类型包装器仅用于 Swagger 文档生成

// NoteSwagger 用于 Swagger 显示 models.Note
type NoteSwagger struct {
	models.Note
}

// UserSwagger 用于 Swagger 显示 models.User
type UserSwagger struct {
	models.User
}

// TagSwagger 用于 Swagger 显示 models.Tag
type TagSwagger struct {
	models.Tag
}
