package handlers

import (
	"errors"
	"strconv"
	"strings"

	"HYH-Blog-Gin/internal/services"
	"HYH-Blog-Gin/internal/utils"

	"github.com/gin-gonic/gin"
)

// NoteHandler 处理笔记相关的请求，封装了笔记业务服务依赖。
type NoteHandler struct {
	svc services.NoteService
}

// NoteCreateRequest 表示创建笔记的请求体。
type NoteCreateRequest struct {
	Title   string   `json:"title" binding:"required"`
	Content string   `json:"content" binding:"required"`
	Tags    []string `json:"tags"`
	Public  *bool    `json:"public"`
}

// NoteUpdateRequest 表示更新笔记的请求体（字段均为可选）。
type NoteUpdateRequest struct {
	Title   *string  `json:"title"`
	Content *string  `json:"content"`
	Tags    []string `json:"tags"`
	Public  *bool    `json:"public"`
}

// NewNoteHandler 创建并返回 NoteHandler 实例（使用 service 层）。
func NewNoteHandler(svc services.NoteService) *NoteHandler {
	return &NoteHandler{svc: svc}
}

// GetNotes 获取当前用户的笔记列表，支持分页。
func (h *NoteHandler) GetNotes(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "unauthorized")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	notes, total, err := h.svc.GetNotes(userID, page, limit)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.Paginated(c, notes, page, limit, total)
}

// CreateNote 创建新笔记，并可同时处理标签。交由 service 完成。
func (h *NoteHandler) CreateNote(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "unauthorized")
		return
	}
	var req NoteCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	note, err := h.svc.CreateNote(userID, strings.TrimSpace(req.Title), req.Content, req.Tags, req.Public)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Created(c, note)
}

// parseUintParam 将字符串解析为 uint，解析失败返回 false。
func parseUintParam(s string) (uint, bool) {
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(v), true
}

// GetNote 获取单个笔记并校验权限（可能公开或作者可访问）。
func (h *NoteHandler) GetNote(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "unauthorized")
		return
	}
	idStr := c.Param("id")
	id, ok := parseUintParam(idStr)
	if !ok {
		utils.BadRequest(c, "invalid id")
		return
	}
	note, err := h.svc.GetNoteByID(userID, id)
	if err != nil {
		if errors.Is(services.ErrNotFound, err) {
			utils.NotFound(c, "note not found")
			return
		}
		if errors.Is(services.ErrForbidden, err) {
			utils.Forbidden(c, "forbidden")
			return
		}
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, note)
}

// UpdateNote 更新笔记内容和标签（交由 service 处理事务与权限）。
func (h *NoteHandler) UpdateNote(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "unauthorized")
		return
	}
	idStr := c.Param("id")
	id, ok := parseUintParam(idStr)
	if !ok {
		utils.BadRequest(c, "invalid id")
		return
	}
	var req NoteUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	note, err := h.svc.UpdateNote(userID, id, req.Title, req.Content, req.Tags, req.Public)
	if err != nil {
		if errors.Is(services.ErrNotFound, err) {
			utils.NotFound(c, "note not found")
			return
		}
		if errors.Is(services.ErrForbidden, err) {
			utils.Forbidden(c, "forbidden")
			return
		}
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, note)
}

// DeleteNote 删除笔记，只允许作者删除。
func (h *NoteHandler) DeleteNote(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "unauthorized")
		return
	}
	idStr := c.Param("id")
	id, ok := parseUintParam(idStr)
	if !ok {
		utils.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.DeleteNote(userID, id); err != nil {
		if errors.Is(services.ErrNotFound, err) {
			utils.NotFound(c, "note not found")
			return
		}
		if errors.Is(services.ErrForbidden, err) {
			utils.Forbidden(c, "forbidden")
			return
		}
		utils.InternalError(c, err.Error())
		return
	}
	utils.NoContent(c)
}
