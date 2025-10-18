package repository

import (
	"HYH-Blog-Gin/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 保证实现关系：若接口变更将在编译期报错
var _ models.TagRepository = (*tagRepository)(nil)

// tagRepository 提供 TagRepository 接口的 GORM 实现。
// 说明：
// - FindOrCreate 支持并发安全创建（基于 ON CONFLICT DO NOTHING），随后统一重查确保主键完整；
// - 通过关联 API（Association）查询某笔记的标签，避免硬编码中间表；
// - 该实现无状态、可在多 goroutine 中复用。
type tagRepository struct{ db *gorm.DB }

func NewTagRepository(db *gorm.DB) models.TagRepository { return &tagRepository{db: db} }

// Create 新建标签
func (r *tagRepository) Create(tag *models.Tag) error { return r.db.Create(tag).Error }

// FindByID 根据主键查询标签
func (r *tagRepository) FindByID(id uint) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.First(&tag, "id = ?", id).Error
	return &tag, err
}

// FindByName 根据名称查询标签
func (r *tagRepository) FindByName(name string) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.First(&tag, "name = ?", name).Error
	return &tag, err
}

// FindOrCreate 批量按名称查找，不存在的按需创建，并保持输入顺序返回，以及每个名称是否为新创建。
func (r *tagRepository) FindOrCreate(names []string) ([]models.Tag, []bool, error) {
	// 去重并保持输入顺序
	order := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, n := range names {
		if _, ok := seen[n]; !ok {
			seen[n] = struct{}{}
			order = append(order, n)
		}
	}
	if len(order) == 0 {
		return []models.Tag{}, []bool{}, nil
	}

	// 查询已存在的标签
	var existing []models.Tag
	if err := r.db.Where("name IN ?", order).Find(&existing).Error; err != nil {
		return nil, nil, err
	}
	byName := make(map[string]models.Tag, len(existing))
	for _, t := range existing {
		byName[t.Name] = t
	}

	// 计算需要创建的标签
	toCreate := make([]models.Tag, 0)
	for _, n := range order {
		if _, ok := byName[n]; !ok {
			toCreate = append(toCreate, models.Tag{Name: n})
		}
	}

	// 并发安全创建：冲突时忽略插入，然后再统一重查
	if len(toCreate) > 0 {
		if err := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&toCreate).Error; err != nil {
			return nil, nil, err
		}
	}

	// 统一重查，确保返回持久化后的主键与字段
	var final []models.Tag
	if err := r.db.Where("name IN ?", order).Find(&final).Error; err != nil {
		return nil, nil, err
	}
	finByName := make(map[string]models.Tag, len(final))
	for _, t := range final {
		finByName[t.Name] = t
	}

	out := make([]models.Tag, 0, len(order))
	created := make([]bool, 0, len(order))
	for _, n := range order {
		if t, ok := finByName[n]; ok {
			out = append(out, t)
			// created is true if the name was NOT present in the initial existing map
			if _, existedBefore := byName[n]; existedBefore {
				created = append(created, false)
			} else {
				created = append(created, true)
			}
		}
	}
	return out, created, nil
}

// FindByNote 查询某笔记的标签集合
func (r *tagRepository) FindByNote(noteID uint) ([]models.Tag, error) {
	// 使用关联 API，而不是手写 JOIN
	note := models.Note{}
	note.ID = noteID
	var tags []models.Tag
	if err := r.db.Model(&note).Association("Tags").Find(&tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// Update 根据主键保存全部字段
func (r *tagRepository) Update(tag *models.Tag) error { return r.db.Save(tag).Error }

// Delete 根据主键删除标签
func (r *tagRepository) Delete(id uint) error {
	return r.db.Delete(&models.Tag{}, "id = ?", id).Error
}

// List 分页列出标签，按 updated_at 降序
func (r *tagRepository) List(page, perPage int) ([]models.Tag, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 50
	}
	var tags []models.Tag
	var total int64
	q := r.db.Model(&models.Tag{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * perPage
	if err := q.Order("updated_at desc").Offset(offset).Limit(perPage).Find(&tags).Error; err != nil {
		return nil, 0, err
	}
	return tags, total, nil
}
