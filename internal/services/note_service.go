package services

import (
	"errors"

	"HYH-Blog-Gin/internal/models"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
)

// NoteService 抽象了笔记相关的业务逻辑。
type NoteService interface {
	GetNotes(userID uint, page, limit int) ([]models.Note, int64, error)
	CreateNote(userID uint, title, content string, tags []string, isPublic *bool) (*models.Note, error)
	GetNoteByID(userID, id uint) (*models.Note, error)
	UpdateNote(userID, id uint, title, content *string, tags []string, isPublic *bool) (*models.Note, error)
	DeleteNote(userID, id uint) error
}

// noteService 是 NoteService 的默认实现，封装 repositories。
type noteService struct {
	notes models.NoteRepository
}

// NewNoteService 创建 NoteService 实例。
func NewNoteService(notes models.NoteRepository) NoteService {
	return &noteService{notes: notes}
}

// GetNotes 分页获取指定用户的笔记列表。
func (s *noteService) GetNotes(userID uint, page, limit int) ([]models.Note, int64, error) {
	return s.notes.FindByAuthor(userID, page, limit)
}

// CreateNote 创建新笔记。
func (s *noteService) CreateNote(userID uint, title, content string, tags []string, isPublic *bool) (*models.Note, error) {
	note := &models.Note{Title: title, Content: content, AuthorID: userID}
	if isPublic != nil {
		note.IsPublic = *isPublic
	}
	if err := s.notes.CreateWithTags(note, tags); err != nil {
		return nil, err
	}
	return note, nil
}

// GetNoteByID 根据 ID 获取笔记，若笔记非公开且非作者则返回 forbidden。
func (s *noteService) GetNoteByID(userID, id uint) (*models.Note, error) {
	note, err := s.notes.FindByID(id)
	if err != nil || note == nil || note.ID == 0 {
		return nil, ErrNotFound
	}
	// 如果笔记不是公开且不是作者，返回 forbidden
	if !note.IsPublic && note.AuthorID != userID {
		return nil, ErrForbidden
	}
	return note, nil
}

// UpdateNote 更新笔记，只有作者可更新。
func (s *noteService) UpdateNote(userID, id uint, title, content *string, tags []string, isPublic *bool) (*models.Note, error) {
	note, err := s.notes.FindByID(id)
	if err != nil || note == nil || note.ID == 0 {
		return nil, ErrNotFound
	}
	if note.AuthorID != userID {
		return nil, ErrForbidden
	}
	if title != nil {
		note.Title = *title
	}
	if content != nil {
		note.Content = *content
	}
	if isPublic != nil {
		note.IsPublic = *isPublic
	}
	if err := s.notes.UpdateWithTags(note, tags); err != nil {
		return nil, err
	}
	return note, nil
}

// DeleteNote 删除笔记，只有作者可删除。
func (s *noteService) DeleteNote(userID, id uint) error {
	note, err := s.notes.FindByID(id)
	if err != nil || note == nil || note.ID == 0 {
		return ErrNotFound
	}
	if note.AuthorID != userID {
		return ErrForbidden
	}
	return s.notes.Delete(id)
}
