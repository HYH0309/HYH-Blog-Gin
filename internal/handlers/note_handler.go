package handlers

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"HYH-Blog-Gin/internal/cache"
	"HYH-Blog-Gin/internal/services"
	"HYH-Blog-Gin/internal/utils"

	"github.com/gin-gonic/gin"
)

// SimpleMessage 通用消息响应
type SimpleMessage struct {
	Message string `json:"message" example:"operation successful"`
}

// NoteHandler 处理笔记相关的请求，封装了笔记业务服务依赖。
type NoteHandler struct {
	svc   services.NoteService
	cache cache.Cache
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
func NewNoteHandler(svc services.NoteService, c cache.Cache) *NoteHandler {
	return &NoteHandler{svc: svc, cache: c}
}

// GetNotes 获取当前用户的笔记列表
// @Summary 获取笔记列表
// @Description 按作者分页获取笔记，返回分页结果（需要鉴权）
// @Tags 笔记
// @Produce json
// @Param page query int false "页码"
// @Param limit query int false "每页数量"
// @Security BearerAuth
// @Success 200 {array} NoteSwagger
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/notes [get]
func (h *NoteHandler) GetNotes(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
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

// CreateNote 创建新笔记
// @Summary 创建笔记
// @Description 创建笔记并可同时处理标签（需要鉴权）
// @Tags 笔记
// @Accept json
// @Produce json
// @Param payload body NoteCreateRequest true "笔记信息"
// @Security BearerAuth
// @Success 201 {object} NoteSwagger
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/notes [post]
func (h *NoteHandler) CreateNote(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
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

// GetNote 获取单个笔记
// @Summary 获取笔记
// @Description 根据 ID 获取单条笔记；如果笔记非公开且非作者则返回 403（需要鉴权）
// @Tags 笔记
// @Produce json
// @Param id path int true "笔记 ID"
// @Security BearerAuth
// @Success 200 {object} NoteSwagger "example: {\"id\":1,\"title\":\"hello\"}"
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/notes/{id} [get]
func (h *NoteHandler) GetNote(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
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

	// 异步增加阅读量（高频写，先写 Redis，再由后台同步至 DB）
	if h.cache != nil {
		go func(id uint) {
			_, err := h.cache.Increment(context.Background(), cache.NewKeyGenerator().NoteViews(id), 1)
			if err != nil {
				return
			}
		}(id)
	}

	utils.OK(c, note)
}

// LikeNote 点赞接口（示例）
// @Summary 给笔记点赞
// @Description 为指定笔记增加一个点赞（高频写，写入 Redis，后台同步到 DB）；需要鉴权
// @Tags 笔记
// @Accept json
// @Produce json
// @Param id path int true "笔记 ID"
// @Security BearerAuth
// @Success 200 {object} SimpleMessage
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/notes/{id}/like [post]
func (h *NoteHandler) LikeNote(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "unauthorized")
		return
	}
	_ = userID // currently unused; could be used to prevent duplicate likes per user
	idStr := c.Param("id")
	id, ok := parseUintParam(idStr)
	if !ok {
		utils.BadRequest(c, "invalid id")
		return
	}
	// 点赞前校验笔记是否存在/可见
	if _, err := h.svc.GetNoteByID(userID, id); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			utils.NotFound(c, "note not found")
			return
		}
		if errors.Is(err, services.ErrForbidden) {
			utils.Forbidden(c, "forbidden")
			return
		}
		utils.InternalError(c, err.Error())
		return
	}
	// 增加 likes 计数（仅写 Redis）
	if h.cache != nil {
		if _, err := h.cache.Increment(context.Background(), cache.NewKeyGenerator().NoteLikes(id), 1); err != nil {
			utils.InternalError(c, "failed to increment like")
			return
		}
	}
	utils.OKMsg(c, "liked", nil)
}

// UpdateNote 更新笔记
// @Summary 更新笔记
// @Description 仅作者可更新笔记，支持部分字段更新并可替换标签集合（需要鉴权）
// @Tags 笔记
// @Accept json
// @Produce json
// @Param id path int true "笔记 ID"
// @Param payload body NoteUpdateRequest true "更新内容（字段可选）"
// @Security BearerAuth
// @Success 200 {object} NoteSwagger
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/notes/{id} [put]
func (h *NoteHandler) UpdateNote(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
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

// DeleteNote 删除笔记
// @Summary 删除笔记
// @Description 仅作者可删除笔记（需要鉴权）
// @Tags 笔记
// @Param id path int true "笔记 ID"
// @Security BearerAuth
// @Success 200 {object} SimpleMessage
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/notes/{id} [delete]
func (h *NoteHandler) DeleteNote(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
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
	utils.OKMsg(c, "note deleted successfully", nil)
}
