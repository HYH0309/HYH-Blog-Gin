package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// WriteFileAtomic 在目标目录创建临时文件，写入内容并尝试重命名到目标路径，保证原子写入。
// 如果目标文件已存在且内容与 data 相同，则不会覆盖并返回 nil。
func WriteFileAtomic(fullPath string, data []byte, perm os.FileMode) error {
	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 在目标目录中创建临时文件
	tmpFile, err := os.CreateTemp(dir, "tmp-*")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpPath := tmpFile.Name()
	// 确保在错误路径上清理临时文件
	closed := false
	defer func() {
		if !closed && tmpFile != nil {
			_ = tmpFile.Close()
		}
		// 成功时不在这里删除，因为重命名会移动它
	}()

	// 写入数据
	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		closed = true
		_ = os.Remove(tmpPath)
		return fmt.Errorf("写入临时文件失败: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		// 尽最大努力同步
	}
	if err := tmpFile.Close(); err != nil {
		// 如果关闭失败，尝试删除临时文件并返回错误
		closed = true
		_ = os.Remove(tmpPath)
		return fmt.Errorf("关闭临时文件失败: %w", err)
	}
	closed = true

	// 尝试重命名
	if err := os.Rename(tmpPath, fullPath); err == nil {
		// 设置目标文件的权限
		_ = os.Chmod(fullPath, perm)
		return nil
	}

	// 重命名失败，可能已存在文件；检查现有文件内容是否相同
	if existing, err := os.ReadFile(fullPath); err == nil {
		s1 := sha256.Sum256(existing)
		s2 := sha256.Sum256(data)
		if hex.EncodeToString(s1[:]) == hex.EncodeToString(s2[:]) {
			// 内容相同；删除临时文件并视为成功
			_ = os.Remove(tmpPath)
			return nil
		}
	}

	// 内容不同或无法读取现有文件：删除临时文件并返回错误
	_ = os.Remove(tmpPath)
	return fmt.Errorf("重命名临时文件到目标路径失败: %w", err)
}
