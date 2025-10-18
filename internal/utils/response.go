package utils

import (
	"net/http"

	"HYH-Blog-Gin/internal/models"

	"github.com/gin-gonic/gin"
)

// JSON 标准响应格式
func JSON(c *gin.Context, status, code int, message string, data interface{}, meta *models.PageMeta) {
	c.JSON(status, models.Response{
		Code:    code,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// OK 返回 200 和数据。
func OK(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, 0, "success", data, nil)
}

// OKMsg 返回 200、消息和数据。
func OKMsg(c *gin.Context, message string, data interface{}) {
	JSON(c, http.StatusOK, 0, message, data, nil)
}

// Created 返回 201 和数据。
func Created(c *gin.Context, data interface{}) {
	JSON(c, http.StatusCreated, 0, "created", data, nil)
}

// NoContent 返回 204 无内容。
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Paginated 返回 200 和分页数据。
func Paginated[T any](c *gin.Context, items []T, page, limit int, total int64) {
	meta := &models.PageMeta{Page: page, Limit: limit, Total: total}
	JSON(c, http.StatusOK, 0, "success", items, meta)
}

// BadRequest 返回 400 错误。
func BadRequest(c *gin.Context, message string) {
	JSON(c, http.StatusBadRequest, http.StatusBadRequest, message, nil, nil)
}

// Unauthorized 返回 401 错误。
func Unauthorized(c *gin.Context, message string) {
	JSON(c, http.StatusUnauthorized, http.StatusUnauthorized, message, nil, nil)
}

// Forbidden 返回 403 错误。
func Forbidden(c *gin.Context, message string) {
	JSON(c, http.StatusForbidden, http.StatusForbidden, message, nil, nil)
}

// NotFound 返回 404 错误。
func NotFound(c *gin.Context, message string) {
	JSON(c, http.StatusNotFound, http.StatusNotFound, message, nil, nil)
}

// Conflict 返回 409 错误。
func Conflict(c *gin.Context, message string) {
	JSON(c, http.StatusConflict, http.StatusConflict, message, nil, nil)
}

// InternalError 返回 500 错误。
func InternalError(c *gin.Context, message string) {
	JSON(c, http.StatusInternalServerError, http.StatusInternalServerError, message, nil, nil)
}

// Unprocessable 返回 422 错误。
func Unprocessable(c *gin.Context, message string) {
	JSON(c, http.StatusUnprocessableEntity, http.StatusUnprocessableEntity, message, nil, nil)
}
