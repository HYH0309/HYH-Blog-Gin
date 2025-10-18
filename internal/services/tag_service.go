package services

import (
	"errors"
	"strings"

	"HYH-Blog-Gin/internal/models"
)

var (
	ErrTagAlreadyExists = errors.New("tag already exists")
	ErrInvalidTagName   = errors.New("invalid tag name")
)

// TagService 提供标签相关业务逻辑。
type TagService interface {
	List(page, perPage int) ([]models.Tag, int64, error)
	Create(name string) (*models.Tag, error)
	GetByID(id uint) (*models.Tag, error)
	Update(id uint, name string) (*models.Tag, error)
	Delete(id uint) error
}

type tagService struct {
	tags models.TagRepository
}

// NewTagService 创建 TagService 实例。
func NewTagService(tags models.TagRepository) TagService {
	return &tagService{tags: tags}
}

func (s *tagService) List(page, perPage int) ([]models.Tag, int64, error) {
	tags, total, err := s.tags.List(page, perPage)
	return tags, total, err
}

func (s *tagService) Create(name string) (*models.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidTagName
	}

	// Use repository's FindOrCreate which returns created flags per name.
	createdTags, createdFlags, err := s.tags.FindOrCreate([]string{name})
	if err != nil {
		return nil, err
	}
	if len(createdTags) == 0 || len(createdFlags) == 0 {
		return nil, nil
	}
	// If the tag already existed (createdFlags[0] == false) we treat it as conflict (409)
	if !createdFlags[0] {
		return &createdTags[0], ErrTagAlreadyExists
	}
	return &createdTags[0], nil
}

func (s *tagService) GetByID(id uint) (*models.Tag, error) {
	return s.tags.FindByID(id)
}

func (s *tagService) Update(id uint, name string) (*models.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidTagName
	}
	t, err := s.tags.FindByID(id)
	if err != nil {
		return nil, err
	}
	// if new name equals old, return unchanged
	if t.Name == name {
		return t, nil
	}
	// ensure no other tag with same name
	if existing, err := s.tags.FindByName(name); err == nil && existing != nil && existing.ID != 0 {
		return nil, ErrTagAlreadyExists
	}

	t.Name = name
	if err := s.tags.Update(t); err != nil {
		// handle unique constraint
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique") || strings.Contains(errStr, "23505") {
			return nil, ErrTagAlreadyExists
		}
		return nil, err
	}
	return t, nil
}

func (s *tagService) Delete(id uint) error {
	return s.tags.Delete(id)
}
