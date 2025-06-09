package main

import (
	"log"
	"sync/atomic"
	"time"

	bonus "github.com/fatorin/mmr-tracker/background"
	"github.com/fatorin/mmr-tracker/config"
	"github.com/fatorin/mmr-tracker/database"
	"github.com/fatorin/mmr-tracker/routes"
	"github.com/gin-gonic/gin"
)

var isProcessing atomic.Bool

func main() {
	config.LoadEnv()
	database.Connect()
	StartBonusWorker()
	r := gin.Default()
	routes.RegisterRoutes(r)

	r.Run(":8080")
}

func StartBonusWorker() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if !isProcessing.CompareAndSwap(false, true) {
				log.Println("⚠️ 尚未完成上一輪，這次跳過")
				continue
			}

			go func() {
				defer isProcessing.Store(false)

				log.Println("🟡 開始處理...")
				start := time.Now()

				ids, err := bonus.GetUnprocessedGameIDs(database.DB)
				if err != nil {
					log.Println("🔴 取得未處理 gameID 失敗:", err)
					return
				}

				for _, id := range ids {
					if err := bonus.ProcessGameBonus(database.DB, id); err != nil {
						log.Printf("❌ Game %d 處理失敗: %v", id, err)
					}
				}

				log.Printf("🟢 完成處理，用時 %s", time.Since(start))
			}()
		}
	}()
}
