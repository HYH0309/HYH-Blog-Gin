package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// GenerateImageSavePath 根据图片 bytes 与原始文件名，生成磁盘保存路径和可访问 URL 路径。
// - baseDir: 本地保存根目录，例如 "./static/images"
// - urlPrefix: 对外访问前缀，例如 "/static/images"
// 返回 urlPath (例如 "/static/images/2025/10/19/abcd1234.webp")，fullPath (例如 "./static/images/2025/10/19/abcd1234.webp")
func GenerateImageSavePath(data []byte, originalFilename, baseDir, urlPrefix string) (string, string, error) {
	// 1. 计算内容哈希（用前 24 hex 字符）
	sum := sha256.Sum256(data)
	hashShort := hex.EncodeToString(sum[:])[:24]

	// 2. 获取并规范扩展名
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext == "" || len(ext) > 8 {
		ext = ".webp" // 默认扩展名
	}

	// 3. 日期目录：年/月/日
	now := time.Now()
	relDir := filepath.Join(fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", now.Month()), fmt.Sprintf("%02d", now.Day()))
	fullDir := filepath.Join(baseDir, relDir)

	// 确保目录存在
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", "", err
	}

	// 3.5: 扫描目录中以 hashShort 开头的现有文件（包括有后缀的情况），若内容相同则直接返回该路径
	if entries, err := os.ReadDir(fullDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.HasPrefix(name, hashShort) {
				p := filepath.Join(fullDir, name)
				if b, err := os.ReadFile(p); err == nil {
					sum2 := sha256.Sum256(b)
					if hex.EncodeToString(sum2[:]) == hex.EncodeToString(sum[:]) {
						urlPath := path.Join(urlPrefix, filepath.ToSlash(relDir), name)
						return urlPath, p, nil
					}
				}
			}
		}
	}

	// 4. 生成基本文件名并处理可能的碰撞：先尝试无后缀的基本名
	filename := hashShort + ext
	fullPath := filepath.Join(fullDir, filename)

	if fi, err := os.Stat(fullPath); err == nil && !fi.IsDir() {
		// 文件存在但内容不同（如果内容相同上面已返回） -> 需要生成带随机后缀的唯一名
		randSuffix, _ := RandHex(6) // 12 hex chars
		filename = fmt.Sprintf("%s-%s%s", hashShort, randSuffix, ext)
		fullPath = filepath.Join(fullDir, filename)
		// 若新生成的名字仍然存在（极小概率），循环直到找到未使用的名字
		for i := 0; i < 8; i++ {
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				break
			}
			randSuffix, _ = RandHex(6)
			filename = fmt.Sprintf("%s-%s%s", hashShort, randSuffix, ext)
			fullPath = filepath.Join(fullDir, filename)
		}
	}

	// 最终 URL path 使用 forward-slash path.Join
	urlPath := path.Join(urlPrefix, filepath.ToSlash(relDir), filename)
	return urlPath, fullPath, nil
}

// RandHex 返回 n 个字节对应的 hex 字符串（长度 = n*2），使用 crypto/rand
func RandHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
