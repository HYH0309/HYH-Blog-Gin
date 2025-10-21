package handlers

import (
	"errors"
	"strconv"

	"HYH-Blog-Gin/internal/services"
	"HYH-Blog-Gin/internal/utils"

	"github.com/gin-gonic/gin"
)

// TagHandler 处理标签相关请求。
type TagHandler struct {
	svc services.TagService
}

// NewTagHandler 创建 TagHandler 实例。
func NewTagHandler(svc services.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

// TagCreateRequest 创建标签请求体
type TagCreateRequest struct {
	Name string `json:"name" binding:"required"`
}

// TagUpdateRequest 更新标签请求体
type TagUpdateRequest struct {
	Name string `json:"name" binding:"required"`
}

// List 列出标签
// @Summary 列出标签
// @Description 分页列出标签（需要鉴权）
// @Tags 标签
// @Produce json
// @Param page query int false "页码"
// @Param per_page query int false "每页数量"
// @Security BearerAuth
// @Success 200 {array} TagSwagger
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/tags [get]
func (h *TagHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 50
	}
	items, total, err := h.svc.List(page, perPage)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.Paginated(c, items, page, perPage, total)
}

// Create 创建标签
// @Summary 创建标签
// @Description 创建新的标签（需要鉴权）。重复创建返回 409
// @Tags 标签
// @Accept json
// @Produce json
// @Param payload body TagCreateRequest true "标签信息"
// @Security BearerAuth
// @Success 201 {object} TagSwagger
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/v1/tags [post]
func (h *TagHandler) Create(c *gin.Context) {
	var req TagCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	tag, err := h.svc.Create(req.Name)
	if err != nil {
		if errors.Is(services.ErrTagAlreadyExists, err) {
			utils.Conflict(c, "tag already exists")
			return
		}
		if errors.Is(services.ErrInvalidTagName, err) {
			utils.BadRequest(c, "invalid tag name")
			return
		}
		utils.InternalError(c, err.Error())
		return
	}
	utils.Created(c, tag)
}

// Get 获取单个标签
// @Summary 获取标签
// @Description 根据 ID 获取标签（需要鉴权）
// @Tags 标签
// @Produce json
// @Param id path int true "标签 ID"
// @Security BearerAuth
// @Success 200 {object} TagSwagger
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/tags/{id} [get]
func (h *TagHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, ok := parseUintParam(idStr)
	if !ok {
		utils.BadRequest(c, "invalid id")
		return
	}
	tag, err := h.svc.GetByID(id)
	if err != nil {
		utils.NotFound(c, "tag not found")
		return
	}
	utils.OK(c, tag)
}

// Update 更新标签
// @Summary 更新标签
// @Description 更新标签名称（需要鉴权）。重复名称返回 409
// @Tags 标签
// @Accept json
// @Produce json
// @Param id path int true "标签 ID"
// @Param payload body TagUpdateRequest true "更新内容"
// @Security BearerAuth
// @Success 200 {object} TagSwagger
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/v1/tags/{id} [put]
func (h *TagHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, ok := parseUintParam(idStr)
	if !ok {
		utils.BadRequest(c, "invalid id")
		return
	}
	var req TagUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	tag, err := h.svc.Update(id, req.Name)
	if err != nil {
		if errors.Is(err, services.ErrTagAlreadyExists) {
			utils.Conflict(c, "tag already exists")
			return
		}
		if err == services.ErrInvalidTagName {
			utils.BadRequest(c, "invalid tag name")
			return
		}
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, tag)
}

// Delete 删除标签
// @Summary 删除标签
// @Description 根据 ID 删除标签（需要鉴权）
// @Tags 标签
// @Param id path int true "标签 ID"
// @Security BearerAuth
// @Success 200 {object} SimpleMessage
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/tags/{id} [delete]
func (h *TagHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, ok := parseUintParam(idStr)
	if !ok {
		utils.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OKMsg(c, "tag deleted", nil)
}
