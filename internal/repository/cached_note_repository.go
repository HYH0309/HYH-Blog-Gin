// Package repository filepath: internal/repository/cached_note_repository.go
package repository

import (
	"context"
	"time"

	"HYH-Blog-Gin/internal/cache"
	"HYH-Blog-Gin/internal/models"
)

// cachedNoteRepository 是 NoteRepository 的包装，提供基于 key 的 Redis 缓存（只用于 FindByID 的示例）。
// 缓存策略：FindByID 读取缓存（JSON），缓存未命中则回退到底层仓储并填充缓存。
// 所有会改变单个笔记数据的写操作（Create/Update/Delete/AddTags/RemoveTags/CreateWithTags/UpdateWithTags）在成功后都会失效对应 key。
// TTL 可配置（构造时传入）。

type cachedNoteRepository struct {
	base  models.NoteRepository
	cache cache.Cache
	ttl   time.Duration
}

// NewCachedNoteRepository 构造一个带缓存的 NoteRepository 包装器。
func NewCachedNoteRepository(base models.NoteRepository, c cache.Cache, ttl time.Duration) models.NoteRepository {
	return &cachedNoteRepository{base: base, cache: c, ttl: ttl}
}

// Create 调用底层创建并在成功后清除缓存（如果有）。
func (r *cachedNoteRepository) Create(note *models.Note) error {
	if err := r.base.Create(note); err != nil {
		return err
	}
	// 删除可能存在的旧缓存（新创建通常不会存在，但保险起见）
	_ = r.cache.Delete(context.Background(), cache.NewKeyGenerator().Note(note.ID))
	return nil
}

func (r *cachedNoteRepository) CreateWithTags(note *models.Note, tagNames []string) error {
	if err := r.base.CreateWithTags(note, tagNames); err != nil {
		return err
	}
	_ = r.cache.Delete(context.Background(), cache.NewKeyGenerator().Note(note.ID))
	return nil
}

func (r *cachedNoteRepository) FindByID(id uint) (*models.Note, error) {
	ctx := context.Background()
	var note models.Note
	key := cache.NewKeyGenerator().Note(id)
	ok, err := r.cache.Get(ctx, key, &note)
	if err != nil {
		// 若缓存读取错误，不阻断，回退到 DB
		return r.base.FindByID(id)
	}
	if ok {
		return &note, nil
	}
	// 缓存未命中
	n, err := r.base.FindByID(id)
	if err != nil {
		return n, err
	}
	// 尝试异步写入缓存：若失败不影响返回（同步写入更简单但可能影响延迟）
	_ = r.cache.Set(ctx, key, n, r.ttl)
	return n, nil
}

func (r *cachedNoteRepository) FindByAuthor(authorID uint, page, limit int) ([]models.Note, int64, error) {
	return r.base.FindByAuthor(authorID, page, limit)
}

func (r *cachedNoteRepository) Search(authorID uint, query string, tags []string) ([]models.Note, error) {
	return r.base.Search(authorID, query, tags)
}

func (r *cachedNoteRepository) Update(note *models.Note) error {
	if err := r.base.Update(note); err != nil {
		return err
	}
	_ = r.cache.Delete(context.Background(), cache.NewKeyGenerator().Note(note.ID))
	return nil
}

func (r *cachedNoteRepository) UpdateWithTags(note *models.Note, tagNames []string) error {
	if err := r.base.UpdateWithTags(note, tagNames); err != nil {
		return err
	}
	_ = r.cache.Delete(context.Background(), cache.NewKeyGenerator().Note(note.ID))
	return nil
}

func (r *cachedNoteRepository) Delete(id uint) error {
	if err := r.base.Delete(id); err != nil {
		return err
	}
	_ = r.cache.Delete(context.Background(), cache.NewKeyGenerator().Note(id))
	return nil
}

func (r *cachedNoteRepository) AddTags(noteID uint, tags []models.Tag) error {
	if err := r.base.AddTags(noteID, tags); err != nil {
		return err
	}
	_ = r.cache.Delete(context.Background(), cache.NewKeyGenerator().Note(noteID))
	return nil
}

func (r *cachedNoteRepository) RemoveTags(noteID uint, tagIDs []uint) error {
	if err := r.base.RemoveTags(noteID, tagIDs); err != nil {
		return err
	}
	_ = r.cache.Delete(context.Background(), cache.NewKeyGenerator().Note(noteID))
	return nil
}
