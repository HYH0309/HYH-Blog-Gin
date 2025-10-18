package models

import (
	"time"

	"gorm.io/gorm"
)

// Image 图片元数据，用于持久化存储文件系统中的图片信息，供快速查询与分页。
type Image struct {
	gorm.Model
	URL     string    `json:"url" gorm:"not null;uniqueIndex:idx_images_url" example:"/static/images/1760854773444000500-de9459314cc6.webp"`
	Path    string    `json:"path" gorm:"not null" example:"D:/data/images/1760854773444000500-de9459314cc6.webp"`
	Size    int64     `json:"size" example:"12345"`
	ModTime time.Time `json:"mod_time" example:"2025-10-19T12:34:56Z"`
}

func (Image) TableName() string { return "images" }

// ImageRepository 定义图片元数据的数据库操作接口。
type ImageRepository interface {
	Create(img *Image) error
	FindByURL(url string) (*Image, error)
	DeleteByURL(url string) error
	List(page, perPage int) ([]Image, int64, error)
}
