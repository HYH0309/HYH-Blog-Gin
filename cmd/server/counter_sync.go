// cmd/server/counter_sync.go
package main

import (
	"context"
	"log"
	"time"

	"HYH-Blog-Gin/internal/cache"
	"HYH-Blog-Gin/internal/models"

	"gorm.io/gorm"
)

// StartCounterSync 启动一个后台 worker，定期把 Redis 中的 views/likes 增量同步回 Postgres。
func StartCounterSync(ctx context.Context, gormDB *gorm.DB, c cache.Cache, interval time.Duration) {
	if c == nil || gormDB == nil {
		log.Println("counter sync: missing dependency, not started")
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	log.Println("counter sync: started")

	for {
		select {
		case <-ctx.Done():
			log.Println("counter sync: stopping")
			return
		case <-ticker.C:
			// do one sync pass
			if err := doSyncOnce(ctx, gormDB, c); err != nil {
				log.Printf("counter sync: pass error: %v", err)
			}
		}
	}
}

// doSyncOnce 执行一次同步任务
func doSyncOnce(ctx context.Context, gormDB *gorm.DB, c cache.Cache) error {
	ids, err := c.PopDirtyNoteIDs(ctx)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	for _, id := range ids {
		// read and clear counters
		views, err := c.GetAndDelete(ctx, cache.NoteViewsKey(id))
		if err != nil {
			// if we cannot read, re-mark id dirty and continue
			log.Printf("counter sync: failed GetAndDelete views for id=%d: %v", id, err)
			_, _ = c.Increment(ctx, cache.NoteViewsKey(id), 0)
			continue
		}
		likes, err := c.GetAndDelete(ctx, cache.NoteLikesKey(id))
		if err != nil {
			log.Printf("counter sync: failed GetAndDelete likes for id=%d: %v", id, err)
			_, _ = c.Increment(ctx, cache.NoteLikesKey(id), 0)
		}

		if views == 0 && likes == 0 {
			continue
		}

		// perform atomic increments using UpdateColumn with gorm.Expr
		tx := gormDB
		err = tx.Transaction(func(tx *gorm.DB) error {
			if views != 0 {
				if err := tx.Model(&models.Note{}).Where("id = ?", id).
					UpdateColumn("views", gorm.Expr("views + ?", views)).Error; err != nil {
					return err
				}
			}
			if likes != 0 {
				if err := tx.Model(&models.Note{}).Where("id = ?", id).
					UpdateColumn("likes", gorm.Expr("likes + ?", likes)).Error; err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("counter sync: failed to update DB for id=%d (views=%d likes=%d): %v", id, views, likes, err)
			// restore counts back to cache so we'll retry later
			if views != 0 {
				if _, ierr := c.Increment(ctx, cache.NoteViewsKey(id), views); ierr != nil {
					log.Printf("counter sync: failed to restore views to redis for id=%d: %v", id, ierr)
				}
			}
			if likes != 0 {
				if _, ierr := c.Increment(ctx, cache.NoteLikesKey(id), likes); ierr != nil {
					log.Printf("counter sync: failed to restore likes to redis for id=%d: %v", id, ierr)
				}
			}
			continue
		}
	}
	return nil
}
