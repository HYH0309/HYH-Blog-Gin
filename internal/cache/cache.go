// Package cache 提供缓存抽象层，支持多种缓存后端实现
// filepath: internal/cache/cache.go
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache 定义缓存抽象接口，支持基本操作和原子计数
type Cache interface {
	// Get 获取缓存值并反序列化到 dest，返回是否存在
	Get(ctx context.Context, key string, dest interface{}) (bool, error)
	// Set 设置缓存值并指定过期时间
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	// Delete 删除指定缓存键
	Delete(ctx context.Context, key string) error
	// Increment 原子增加指定键的值，返回操作后的值
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	// GetInteger 获取整数值，不存在时返回0
	GetInteger(ctx context.Context, key string) (int64, error)
	// GetAndDelete 原子获取并删除键值，用于获取增量后清零
	GetAndDelete(ctx context.Context, key string) (int64, error)
	// PopDirtyNoteIDs 获取并清空脏笔记ID集合
	PopDirtyNoteIDs(ctx context.Context) ([]uint, error)
}

// 缓存键常量定义
const (
	// DirtyNoteSetKey 脏笔记ID集合键名
	DirtyNoteSetKey = "note:counters:dirty"
	// KeyPrefixNote 笔记缓存键前缀
	KeyPrefixNote = "note:"
	// KeySuffixViews 浏览量计数器后缀
	KeySuffixViews = "views"
	// KeySuffixLikes 点赞数计数器后缀
	KeySuffixLikes = "likes"
)

// KeyGenerator 缓存键生成器
type KeyGenerator struct{}

// NewKeyGenerator 创建缓存键生成器实例
func NewKeyGenerator() *KeyGenerator {
	return &KeyGenerator{}
}

// Note 生成笔记缓存键
func (kg *KeyGenerator) Note(id uint) string {
	return fmt.Sprintf("%s%d", KeyPrefixNote, id)
}

// NoteViews 生成笔记浏览量计数器键
func (kg *KeyGenerator) NoteViews(id uint) string {
	return fmt.Sprintf("%s%d:%s", KeyPrefixNote, id, KeySuffixViews)
}

// NoteLikes 生成笔记点赞数计数器键
func (kg *KeyGenerator) NoteLikes(id uint) string {
	return fmt.Sprintf("%s%d:%s", KeyPrefixNote, id, KeySuffixLikes)
}

// RedisCache 基于Redis的缓存实现
type RedisCache struct {
	client *redis.Client
	keys   *KeyGenerator
}

// NewRedisCache 创建Redis缓存实例
func NewRedisCache(client *redis.Client) Cache {
	return &RedisCache{
		client: client,
		keys:   NewKeyGenerator(),
	}
}

// Set 设置缓存值
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	serializedValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化缓存值失败: %w", err)
	}

	if err := rc.client.Set(ctx, key, serializedValue, ttl).Err(); err != nil {
		return fmt.Errorf("设置缓存失败: %w", err)
	}

	return nil
}

// Get 获取缓存值
func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	result, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(redis.Nil, err) {
			return false, nil
		}
		return false, fmt.Errorf("获取缓存失败: %w", err)
	}

	if err := json.Unmarshal([]byte(result), dest); err != nil {
		return false, fmt.Errorf("反序列化缓存值失败: %w", err)
	}

	return true, nil
}

// Delete 删除缓存键
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	if err := rc.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("删除缓存键失败: %w", err)
	}
	return nil
}

// Increment 原子增加计数器值并标记脏数据
func (rc *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	newValue, err := rc.client.IncrBy(ctx, key, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("增加计数器失败: %w", err)
	}

	// 标记对应的笔记ID为脏数据
	if noteID, isValid := rc.extractNoteIDFromCounterKey(key); isValid {
		rc.markNoteAsDirty(ctx, noteID)
	}

	return newValue, nil
}

// GetInteger 获取整数值
func (rc *RedisCache) GetInteger(ctx context.Context, key string) (int64, error) {
	result, err := rc.client.Get(ctx, key).Int64()
	if err != nil {
		if errors.Is(redis.Nil, err) {
			return 0, nil
		}
		return 0, fmt.Errorf("获取整数值失败: %w", err)
	}
	return result, nil
}

// GetAndDelete 原子获取并删除整数值
func (rc *RedisCache) GetAndDelete(ctx context.Context, key string) (int64, error) {
	// 使用 Redis 6.2+ 的 GETDEL 命令
	result, err := rc.client.GetDel(ctx, key).Result()
	if err != nil {
		if errors.Is(redis.Nil, err) {
			return 0, nil
		}
		return 0, fmt.Errorf("获取并删除键值失败: %w", err)
	}

	if result == "" {
		return 0, nil
	}

	value, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("解析整数值失败: %w", err)
	}

	return value, nil
}

// PopDirtyNoteIDs 获取并清空脏笔记ID集合
func (rc *RedisCache) PopDirtyNoteIDs(ctx context.Context) ([]uint, error) {
	members, err := rc.client.SMembers(ctx, DirtyNoteSetKey).Result()
	if err != nil {
		if errors.Is(redis.Nil, err) {
			return nil, nil
		}
		return nil, fmt.Errorf("获取脏笔记集合失败: %w", err)
	}

	if len(members) == 0 {
		return nil, nil
	}

	// 清空脏数据集合
	if err := rc.client.Del(ctx, DirtyNoteSetKey).Err(); err != nil {
		return nil, fmt.Errorf("清空脏笔记集合失败: %w", err)
	}

	// 转换字符串ID为uint类型
	noteIDs := make([]uint, 0, len(members))
	for _, member := range members {
		if id, err := strconv.ParseUint(member, 10, 64); err == nil {
			noteIDs = append(noteIDs, uint(id))
		}
	}

	return noteIDs, nil
}

// extractNoteIDFromCounterKey 从计数器键名中提取笔记ID
func (rc *RedisCache) extractNoteIDFromCounterKey(key string) (uint, bool) {
	if !strings.HasPrefix(key, KeyPrefixNote) {
		return 0, false
	}

	keyParts := strings.Split(key, ":")
	if len(keyParts) != 3 {
		return 0, false
	}

	// 验证计数器类型
	if keyParts[2] != KeySuffixViews && keyParts[2] != KeySuffixLikes {
		return 0, false
	}

	noteID, err := strconv.ParseUint(keyParts[1], 10, 64)
	if err != nil {
		return 0, false
	}

	return uint(noteID), true
}

// markNoteAsDirty 标记笔记为脏数据
func (rc *RedisCache) markNoteAsDirty(ctx context.Context, noteID uint) {
	// 使用异步操作，不阻塞主流程
	member := strconv.FormatUint(uint64(noteID), 10)
	go func() {
		_ = rc.client.SAdd(ctx, DirtyNoteSetKey, member).Err()
	}()
}

// NoOpCache 无操作缓存实现，用于测试或禁用缓存场景
type NoOpCache struct{}

// NewNoOpCache 创建无操作缓存实例
func NewNoOpCache() Cache {
	return &NoOpCache{}
}

// Get 无操作实现
func (noc *NoOpCache) Get(_ context.Context, _ string, _ interface{}) (bool, error) {
	return false, nil
}

// Set 无操作实现
func (noc *NoOpCache) Set(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
	return nil
}

// Delete 无操作实现
func (noc *NoOpCache) Delete(_ context.Context, _ string) error {
	return nil
}

// Increment 无操作实现
func (noc *NoOpCache) Increment(_ context.Context, _ string, _ int64) (int64, error) {
	return 0, nil
}

// GetInteger 无操作实现
func (noc *NoOpCache) GetInteger(_ context.Context, _ string) (int64, error) {
	return 0, nil
}

// GetAndDelete 无操作实现
func (noc *NoOpCache) GetAndDelete(_ context.Context, _ string) (int64, error) {
	return 0, nil
}

// PopDirtyNoteIDs 无操作实现
func (noc *NoOpCache) PopDirtyNoteIDs(_ context.Context) ([]uint, error) {
	return nil, nil
}
