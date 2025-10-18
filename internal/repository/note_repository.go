package repository

import (
	"HYH-Blog-Gin/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 保证实现关系：若接口变更将在编译期报错
var _ models.NoteRepository = (*noteRepository)(nil)

// noteRepository 提供 NoteRepository 接口的 GORM 实现。
// 特性：
// - 通过 Preload 预加载 Author 与 Tags，减少 N+1 查询。
// - 在标签关联增删时，使用仅含主键的 stub 实体避免额外 SELECT。
// - 分页参数健壮性处理：当 limit<=0 时不应用分页，返回全部结果。
// - 所有方法均是轻量无状态实现，可在多 goroutine 中共享。
type noteRepository struct{ db *gorm.DB }

// NewNoteRepository 构造一个基于 GORM 的笔记仓储实现。
func NewNoteRepository(db *gorm.DB) models.NoteRepository { return &noteRepository{db: db} }

// Create 新建笔记记录。
func (r *noteRepository) Create(note *models.Note) error { return r.db.Create(note).Error }

// CreateWithTags 在单个事务中创建笔记并处理标签关联（保证原子性）。
func (r *noteRepository) CreateWithTags(note *models.Note, tagNames []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(note).Error; err != nil {
			return err
		}
		if len(tagNames) == 0 {
			return nil
		}
		// 去重并保持顺序
		order := make([]string, 0, len(tagNames))
		seen := make(map[string]struct{}, len(tagNames))
		for _, n := range tagNames {
			if _, ok := seen[n]; !ok {
				seen[n] = struct{}{}
				order = append(order, n)
			}
		}
		// 准备要创建的标签实体
		var toCreate []models.Tag
		for _, n := range order {
			toCreate = append(toCreate, models.Tag{Name: n})
		}
		if len(toCreate) > 0 {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&toCreate).Error; err != nil {
				return err
			}
		}
		// 重新查询以获取完整主键
		var final []models.Tag
		if err := tx.Where("name IN ?", order).Find(&final).Error; err != nil {
			return err
		}
		// 关联标签
		if err := tx.Model(note).Association("Tags").Append(&final); err != nil {
			return err
		}
		note.Tags = final
		return nil
	})
}

// FindByID 根据主键查询笔记，预加载作者与标签。
func (r *noteRepository) FindByID(id uint) (*models.Note, error) {
	var note models.Note
	err := r.db.Preload("Author").Preload("Tags").First(&note, "id = ?", id).Error
	return &note, err
}

// FindByAuthor 按作者分页查询笔记，返回列表与总数。
// 当 limit<=0 时，不应用分页（返回全部）。
func (r *noteRepository) FindByAuthor(authorID uint, page, limit int) ([]models.Note, int64, error) {
	var notes []models.Note
	var total int64

	qCount := r.db.Model(&models.Note{}).Where("author_id = ?", authorID)
	if err := qCount.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	db := r.db.Preload("Author").Preload("Tags").Where("author_id = ?", authorID).Order("notes.created_at DESC")
	if limit > 0 {
		if page <= 0 {
			page = 1
		}
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	err := db.Find(&notes).Error
	return notes, total, err
}

// Search 在作者空间内按标题/内容与标签过滤，并按创建时间倒序返回。
// - query 支持 ILIKE（PostgresSQL）模糊匹配。
// - tags 非空时基于关联 Join 过滤并去重。
func (r *noteRepository) Search(authorID uint, query string, tags []string) ([]models.Note, error) {
	var notes []models.Note
	// 基础条件 + 预加载
	db := r.db.Model(&models.Note{}).Preload("Author").Preload("Tags").Where("author_id = ?", authorID)
	if query != "" {
		q := "%" + query + "%"
		db = db.Where("title ILIKE ? OR content ILIKE ?", q, q)
	}
	// 若根据标签过滤：使用关联名进行 Join，避免手写中间表名；并使用 Distinct 去重
	if len(tags) > 0 {
		db = db.Joins("Tags").Where("tags.name IN ?", tags).Distinct("notes.*")
	}
	err := db.Order("notes.created_at DESC").Find(&notes).Error
	return notes, err
}

// Update 根据主键保存全部字段。
func (r *noteRepository) Update(note *models.Note) error { return r.db.Save(note).Error }

// UpdateWithTags 在单个事务中更新笔记并替换标签集合（保证原子性）。
func (r *noteRepository) UpdateWithTags(note *models.Note, tagNames []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 保存 note 本体
		if err := tx.Save(note).Error; err != nil {
			return err
		}
		// 如果 tagNames 为 nil，表示不改变标签集合
		if tagNames == nil {
			return nil
		}
		// 如果为空数组则清空关联
		if len(tagNames) == 0 {
			if err := tx.Model(note).Association("Tags").Clear(); err != nil {
				return err
			}
			return nil
		}
		// 去重并保持顺序
		order := make([]string, 0, len(tagNames))
		seen := make(map[string]struct{}, len(tagNames))
		for _, n := range tagNames {
			if _, ok := seen[n]; !ok {
				seen[n] = struct{}{}
				order = append(order, n)
			}
		}
		// 创建缺失的标签（并发安全）
		var toCreate []models.Tag
		for _, n := range order {
			toCreate = append(toCreate, models.Tag{Name: n})
		}
		if len(toCreate) > 0 {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&toCreate).Error; err != nil {
				return err
			}
		}
		// 取得最终标签集合
		var final []models.Tag
		if err := tx.Where("name IN ?", order).Find(&final).Error; err != nil {
			return err
		}
		// 使用 Replace 保证替换旧关联为新集合
		if err := tx.Model(note).Association("Tags").Replace(&final); err != nil {
			return err
		}
		note.Tags = final
		return nil
	})
}

// Delete 根据主键删除笔记（软删除）。
func (r *noteRepository) Delete(id uint) error {
	return r.db.Delete(&models.Note{}, "id = ?", id).Error
}

// AddTags 为指定笔记追加标签集合。避免先 SELECT 笔记，直接使用仅含主键的 stub 实体。
func (r *noteRepository) AddTags(noteID uint, tags []models.Tag) error {
	if len(tags) == 0 {
		return nil
	}
	note := models.Note{}
	note.ID = noteID
	return r.db.Model(&note).Association("Tags").Append(tags)
}

// RemoveTags 从指定笔记移除一组标签（按标签 ID）。
// 若 tagIDs 为空，则不做任何操作。
func (r *noteRepository) RemoveTags(noteID uint, tagIDs []uint) error {
	if len(tagIDs) == 0 {
		return nil
	}
	// Association.Delete 需要传入待移除的实体列表，而不是 where 条件
	var toDelete []models.Tag
	if err := r.db.Where("id IN ?", tagIDs).Find(&toDelete).Error; err != nil {
		return err
	}
	note := models.Note{}
	note.ID = noteID
	return r.db.Model(&note).Association("Tags").Delete(&toDelete)
}
