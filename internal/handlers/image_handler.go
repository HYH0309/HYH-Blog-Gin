package handlers

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"HYH-Blog-Gin/internal/services"
	"HYH-Blog-Gin/internal/utils"

	"github.com/gin-gonic/gin"
)

// ImageHandler 提供图片上传、列表、查询和删除的 HTTP 处理逻辑。
type ImageHandler struct {
	svc services.ImageService
}

// NewImageHandler 创建 ImageHandler
func NewImageHandler(svc services.ImageService) *ImageHandler {
	return &ImageHandler{svc: svc}
}

// ImageUploadResponse 上传图片成功返回的结构
type ImageUploadResponse struct {
	URL string `json:"url" example:"/static/images/1760854773444000500-de9459314cc6.webp"`
}

// ImageListResponse 图片列表响应（带分页）
type ImageListResponse struct {
	Total int                  `json:"total" example:"2"`
	Items []services.ImageMeta `json:"items"`
}

// SimpleBoolResponse 通用布尔响应，例如删除操作
type SimpleBoolResponse struct {
	OK bool `json:"ok" example:"true"`
}

// Upload 上传图片
// @Summary 上传图片
// @Description 上传文件并通过图片转换服务生成 webp，返回可访问的 URL（需要鉴权）
// @Tags 图片
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "上传的文件"
// @Param filename formData string false "可选的文件名（包含扩展名），不提供将使用原始文件名"
// @Security BearerAuth
// @Success 200 {object} ImageUploadResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 502 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/images [post]
func (h *ImageHandler) Upload(c *gin.Context) {
	const maxUploadSize = 10 << 20 // 10 MiB

	// Protect request body size
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)
	if err := c.Request.ParseMultipartForm(maxUploadSize); err != nil {
		utils.BadRequest(c, "multipart form parse error or file too large")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "file 为必填字段")
		return
	}

	// quick size check using header if available
	if fileHeader.Size > maxUploadSize {
		utils.BadRequest(c, "file size exceeds limit")
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		log.Printf("打开上传文件失败: %v", err)
		utils.InternalError(c, "读取上传文件失败")
		return
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			log.Printf("关闭上传文件时出错: %v", cerr)
		}
	}()

	data, err := io.ReadAll(f)
	if err != nil {
		log.Printf("读取上传文件失败: %v", err)
		utils.InternalError(c, "读取上传文件失败")
		return
	}

	// MIME sniff (first 512 bytes)
	sniffLen := 512
	if len(data) < sniffLen {
		sniffLen = len(data)
	}
	contentType := http.DetectContentType(data[:sniffLen])
	if !utils.IsAllowedImageContentType(contentType) {
		utils.BadRequest(c, "unsupported file type: "+contentType)
		return
	}

	origName := c.PostForm("filename")
	if origName == "" {
		origName = fileHeader.Filename
	}
	// sanitize filename to avoid path traversal & unsafe chars
	origName = utils.SanitizeFileName(origName)

	// pass to service (service will ensure .webp extension if needed)
	url, err := h.svc.Save(data, origName)
	if err != nil {
		log.Printf("保存图片失败: %v", err)
		// 如果是上游转换服务不可用或转换失败，返回 502 Bad Gateway
		if strings.Contains(err.Error(), "image conversion service unavailable") || strings.Contains(err.Error(), "remote image conversion failed") {
			utils.JSON(c, http.StatusBadGateway, http.StatusBadGateway, err.Error(), nil, nil)
			return
		}
		utils.InternalError(c, "保存图片失败")
		return
	}

	utils.OK(c, ImageUploadResponse{URL: url})
}

// List 列出图片，支持分页
// @Summary 列出图片
// @Description 返回图片分页列表（需要鉴权）
// @Tags 图片
// @Produce json
// @Param page query int false "页码"
// @Param per_page query int false "每页数量"
// @Security BearerAuth
// @Success 200 {object} ImageListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/images [get]
func (h *ImageHandler) List(c *gin.Context) {
	page := 1
	perPage := 50
	if p := c.Query("page"); p != "" {
		if pv, err := strconv.Atoi(p); err == nil && pv > 0 {
			page = pv
		}
	}
	if pp := c.Query("per_page"); pp != "" {
		if pv, err := strconv.Atoi(pp); err == nil && pv > 0 {
			perPage = pv
		}
	}

	items, total, err := h.svc.ListImages(page, perPage)
	if err != nil {
		log.Printf("列出图片失败: %v", err)
		utils.InternalError(c, "列出图片失败")
		return
	}

	utils.OK(c, ImageListResponse{Total: total, Items: items})
}

// Info 返回图片元信息
// @Summary 图片信息
// @Description 根据 urlPath 返回单张图片的元信息（需要鉴权）
// @Tags 图片
// @Produce json
// @Param url query string true "图片 URL 路径，例如 /static/images/..."
// @Security BearerAuth
// @Success 200 {object} services.ImageMeta
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/images/info [get]
func (h *ImageHandler) Info(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		utils.BadRequest(c, "url 参数必填")
		return
	}
	meta, err := h.svc.GetImageInfo(url)
	if err != nil {
		utils.NotFound(c, "image not found")
		return
	}
	utils.OK(c, meta)
}

// Delete 删除图片
// @Summary 删除图片
// @Description 删除指定图片文件及其元数据（需要鉴权）
// @Tags 图片
// @Param url query string true "图片 URL 路径，例如 /static/images..."
// @Security BearerAuth
// @Success 200 {object} SimpleBoolResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/images [delete]
func (h *ImageHandler) Delete(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		utils.BadRequest(c, "url 参数必填")
		return
	}
	if err := h.svc.DeleteImage(url); err != nil {
		utils.InternalError(c, "删除失败")
		return
	}
	utils.OK(c, SimpleBoolResponse{OK: true})
}
