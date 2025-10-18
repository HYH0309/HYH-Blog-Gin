package repository

import (
	"HYH-Blog-Gin/internal/models"

	"gorm.io/gorm"
)

// imageRepository 提供 ImageRepository 接口的 GORM 实现。
type imageRepository struct{ db *gorm.DB }

// NewImageRepository 构造基于 GORM 的图片元数据仓储实现。
func NewImageRepository(db *gorm.DB) models.ImageRepository { return &imageRepository{db: db} }

func (r *imageRepository) Create(img *models.Image) error {
	return r.db.Create(img).Error
}

func (r *imageRepository) FindByURL(url string) (*models.Image, error) {
	var img models.Image
	err := r.db.First(&img, "url = ?", url).Error
	return &img, err
}

func (r *imageRepository) DeleteByURL(url string) error {
	return r.db.Delete(&models.Image{}, "url = ?", url).Error
}

func (r *imageRepository) List(page, perPage int) ([]models.Image, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 50
	}
	var imgs []models.Image
	var total int64
	q := r.db.Model(&models.Image{})
	q.Count(&total)
	offset := (page - 1) * perPage
	err := q.Order("updated_at desc").Offset(offset).Limit(perPage).Find(&imgs).Error
	return imgs, total, err
}
