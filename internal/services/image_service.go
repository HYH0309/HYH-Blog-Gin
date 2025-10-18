package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"HYH-Blog-Gin/internal/models"
	"HYH-Blog-Gin/proto/imageconvpb"
)

// ImageMeta 描述存储在文件系统中的图片的元信息。
type ImageMeta struct {
	URL     string    `json:"url"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

// ImageService 提供图片保存（通过 gRPC 转为 webp）、图片的增删查列表操作。
type ImageService interface {
	// Save 上传并通过 gRPC 转为 webp（quality 1-100 可在实现中调整），返回可访问 URL
	Save(data []byte, originalFilename string) (string, error)
	// ListImages 返回按修改时间倒序的图片元信息，支持分页（page 从 1 开始）
	ListImages(page, perPage int) ([]ImageMeta, int, error)
	// GetImageInfo 基于 urlPath 返回元信息
	GetImageInfo(urlPath string) (ImageMeta, error)
	// DeleteImage 删除指定 urlPath 的文件
	DeleteImage(urlPath string) error
}

type imageService struct {
	client  imageconvpb.ImageServiceClient // required: use gRPC conversion
	storage Storage
	quality int                    // default quality for conversion
	imgRepo models.ImageRepository // optional: repository for metadata persistence
}

// NewImageService 创建 ImageService，storage 为可替换的存储后端（本地或 OSS）；imgRepo 可为 nil 表示不使用 DB
func NewImageService(client imageconvpb.ImageServiceClient, storage Storage, quality int, imgRepo models.ImageRepository) ImageService {
	if quality <= 0 || quality > 100 {
		quality = 80
	}
	return &imageService{client: client, storage: storage, quality: quality, imgRepo: imgRepo}
}

// Save 实现：必须使用 gRPC 转换；若 client 为空或转换失败，则返回错误，不保存原图。
func (s *imageService) Save(data []byte, originalFilename string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("image conversion service unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	reply, err := s.client.ConvertImage(ctx, &imageconvpb.ImageRequest{Data: data, Quality: int32(s.quality)})
	if err != nil {
		return "", fmt.Errorf("remote image conversion failed: %w", err)
	}
	if reply == nil || len(reply.Data) == 0 {
		return "", fmt.Errorf("remote image conversion returned empty result")
	}
	webpData := reply.Data

	// ensure filename extension is .webp
	orig := originalFilename
	ext := strings.ToLower(filepath.Ext(orig))
	if ext != ".webp" {
		orig = strings.TrimSuffix(orig, ext) + ".webp"
	}

	urlPath, fullPath, err := s.storage.Save(webpData, orig)
	if err != nil {
		return "", err
	}

	// 如果配置了仓储，尝试将元数据写入数据库；若失败，则删除已写入的文件并返回错误
	if s.imgRepo != nil {
		meta, _ := s.storage.GetImageInfo(urlPath) // best-effort
		img := &models.Image{URL: urlPath, Path: fullPath, Size: meta.Size, ModTime: meta.ModTime}
		if err := s.imgRepo.Create(img); err != nil {
			// 回滚：尝试删除文件（忽略删除错误，但返回原始 DB 错误）
			_ = s.storage.Delete(urlPath)
			return "", fmt.Errorf("save metadata to db: %w", err)
		}
	}
	return urlPath, nil
}

// ListImages 首选从 DB 查询，回退到 storage
func (s *imageService) ListImages(page, perPage int) ([]ImageMeta, int, error) {
	if s.imgRepo != nil {
		imgs, total, err := s.imgRepo.List(page, perPage)
		if err == nil {
			metas := make([]ImageMeta, 0, len(imgs))
			for _, im := range imgs {
				metas = append(metas, ImageMeta{URL: im.URL, Path: im.Path, Size: im.Size, ModTime: im.ModTime})
			}
			return metas, int(total), nil
		}
		// 若 DB 查询错误，回退到 storage
	}
	return s.storage.ListImages(page, perPage)
}

// GetImageInfo 首选 DB，再回退 storage
func (s *imageService) GetImageInfo(urlPath string) (ImageMeta, error) {
	if s.imgRepo != nil {
		img, err := s.imgRepo.FindByURL(urlPath)
		if err == nil {
			return ImageMeta{URL: img.URL, Path: img.Path, Size: img.Size, ModTime: img.ModTime}, nil
		}
		// 回退到 storage
	}
	return s.storage.GetImageInfo(urlPath)
}

// DeleteImage 同步删除文件与 DB（先删文件，后删 DB）
func (s *imageService) DeleteImage(urlPath string) error {
	// 先删除文件
	if err := s.storage.Delete(urlPath); err != nil {
		return err
	}
	// 再删除 DB 记录（如果存在）
	if s.imgRepo != nil {
		if err := s.imgRepo.DeleteByURL(urlPath); err != nil {
			return err
		}
	}
	return nil
}
