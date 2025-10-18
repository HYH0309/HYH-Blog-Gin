package services

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"HYH-Blog-Gin/internal/utils"
)

// Storage 定义图片存储后端接口。
type Storage interface {
	Save(data []byte, originalFilename string) (urlPath string, fullPath string, err error)
	GetImageInfo(urlPath string) (ImageMeta, error)
	Delete(urlPath string) error
	ListImages(page, perPage int) ([]ImageMeta, int, error)
}

// LocalStorage implements Storage 使用本地文件系统存储图片。
type LocalStorage struct {
	saveDir string
	urlBase string
}

func NewLocalStorage(saveDir, urlBase string) *LocalStorage {
	return &LocalStorage{saveDir: saveDir, urlBase: urlBase}
}

func (ls *LocalStorage) Save(data []byte, originalFilename string) (string, string, error) {
	urlPath, fullPath, err := utils.GenerateImageSavePath(data, originalFilename, ls.saveDir, ls.urlBase)
	if err != nil {
		return "", "", fmt.Errorf("generate save path: %w", err)
	}
	if err := utils.WriteFileAtomic(fullPath, data, 0644); err != nil {
		return "", "", fmt.Errorf("write file: %w", err)
	}
	return urlPath, fullPath, nil
}

func (ls *LocalStorage) GetImageInfo(urlPath string) (ImageMeta, error) {
	if !strings.HasPrefix(urlPath, ls.urlBase) {
		return ImageMeta{}, fmt.Errorf("url does not have expected prefix")
	}
	rel := strings.TrimPrefix(urlPath, ls.urlBase)
	rel = strings.TrimPrefix(rel, "/")
	if strings.Contains(rel, "..") {
		return ImageMeta{}, fmt.Errorf("invalid path")
	}
	full := filepath.Join(ls.saveDir, filepath.FromSlash(rel))
	fi, err := os.Stat(full)
	if err != nil {
		return ImageMeta{}, err
	}
	return ImageMeta{URL: path.Join(ls.urlBase, filepath.ToSlash(rel)), Path: full, Size: fi.Size(), ModTime: fi.ModTime()}, nil
}

func (ls *LocalStorage) Delete(urlPath string) error {
	if !strings.HasPrefix(urlPath, ls.urlBase) {
		return fmt.Errorf("url does not have expected prefix")
	}
	rel := strings.TrimPrefix(urlPath, ls.urlBase)
	rel = strings.TrimPrefix(rel, "/")
	if strings.Contains(rel, "..") {
		return fmt.Errorf("invalid path")
	}
	full := filepath.Join(ls.saveDir, filepath.FromSlash(rel))
	return os.Remove(full)
}

func (ls *LocalStorage) ListImages(page, perPage int) ([]ImageMeta, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 50
	}
	var metas []ImageMeta
	err := filepath.Walk(ls.saveDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(ls.saveDir, p)
		if rerr != nil {
			return nil
		}
		url := path.Join(ls.urlBase, filepath.ToSlash(rel))
		metas = append(metas, ImageMeta{URL: url, Path: p, Size: info.Size(), ModTime: info.ModTime()})
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(metas, func(i, j int) bool {
		return metas[i].ModTime.After(metas[j].ModTime)
	})
	total := len(metas)
	start := (page - 1) * perPage
	if start >= total {
		return []ImageMeta{}, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}
	return metas[start:end], total, nil
}
