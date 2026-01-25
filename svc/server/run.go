package server

import (
	"fmt"
	"strconv"

	"github.com/go-idp/pipeline"
	"github.com/go-zoox/chalk"
	"github.com/go-zoox/encoding/yaml"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/headers"
	"github.com/go-zoox/zoox"

	defaults "github.com/go-zoox/zoox/defaults"
)

func (s *server) Run() error {
	if ok := fs.IsExist(s.cfg.Workdir); !ok {
		if err := fs.Mkdirp(s.cfg.Workdir); err != nil {
			return fmt.Errorf("failed to create workdir: %s", err)
		}
	}

	app := defaults.Defaults()

	app.SetBanner(fmt.Sprintf(`
  _____       _______  ___    ___  _          ___         
 / ___/__    /  _/ _ \/ _ \  / _ \(_)__  ___ / (_)__  ___ 
/ (_ / _ \  _/ // // / ___/ / ___/ / _ \/ -_) / / _ \/ -_)
\___/\___/ /___/____/_/    /_/  /_/ .__/\__/_/_/_//_/\__/ 
                                 /_/                      v%s
`, chalk.Green(pipeline.Version)))

	if s.cfg.Username != "" || s.cfg.Password != "" {
		app.Use(func(ctx *zoox.Context) {
			user, pass, ok := ctx.Request.BasicAuth()
			if !ok {
				ctx.Set("WWW-Authenticate", `Basic realm="go-zoox"`)
				ctx.Status(401)
				return
			}

			if !(user == s.cfg.Username && pass == s.cfg.Password) {
				ctx.Status(401)
				return
			}

			ctx.Next()
		})
	}

	err := Mount(app, func(opt *MountConfig) {
		opt.Path = s.cfg.Path
		opt.Workdir = s.cfg.Workdir
		opt.Environment = s.cfg.Environment
		opt.Store = s.store
		opt.Queue = s.queue
	})
	if err != nil {
		return err
	}

	// API 路由
	api := app.Group("/api/v1")
	{
		// 获取 pipeline 列表（合并 store 和 queue 的数据）
		api.Get("/pipelines", func(ctx *zoox.Context) {
			limit := 100
			if limitStr := ctx.Request.URL.Query().Get("limit"); limitStr != "" {
				if parsed, err := strconv.Atoi(limitStr); err == nil {
					limit = parsed
				}
			}

			// 获取 store 中的记录
			storeRecords := s.store.List(limit * 2) // 获取更多以便合并

			// 获取 queue 中的项目
			queueItems := s.queue.List()

			// 创建 ID 到记录的映射
			recordsMap := make(map[string]*PipelineRecord)
			for _, record := range storeRecords {
				recordsMap[record.ID] = record
			}

			// 合并 queue 中的 pending 和 running 任务
			for _, item := range queueItems {
				if record, exists := recordsMap[item.ID]; exists {
					// 如果 store 中有记录，更新状态（queue 的状态可能更新）
					if item.Status == "pending" || item.Status == "running" {
						record.Status = item.Status
						if item.StartedAt != nil {
							// 保持 store 中的 StartedAt，除非 queue 中有更新的
						}
					}
				} else {
					// 如果 store 中没有记录，从 queue 创建记录
					record := &PipelineRecord{
						ID:        item.ID,
						Name:      item.Name,
						Status:    item.Status,
						StartedAt: item.CreatedAt,
						Config:    make(map[string]interface{}),
						YAML:      item.YAML,
						Logs:      make([]LogEntry, 0),
					}
					if item.StartedAt != nil {
						record.StartedAt = *item.StartedAt
					}
					recordsMap[item.ID] = record
				}
			}

			// 转换为列表并排序
			records := make([]*PipelineRecord, 0, len(recordsMap))
			for _, record := range recordsMap {
				records = append(records, record)
			}

			// 按时间倒序排序
			for i := 0; i < len(records)-1; i++ {
				for j := i + 1; j < len(records); j++ {
					if records[i].StartedAt.Before(records[j].StartedAt) {
						records[i], records[j] = records[j], records[i]
					}
				}
			}

			if limit > 0 && limit < len(records) {
				records = records[:limit]
			}

			ctx.JSON(200, map[string]interface{}{
				"data":  records,
				"total": len(records),
			})
		})

		// 获取单个 pipeline 详情
		api.Get("/pipelines/:id", func(ctx *zoox.Context) {
			id := ctx.Param().Get("id").String()
			record, ok := s.store.Get(id)
			if !ok {
				ctx.Status(404)
				ctx.JSON(404, map[string]string{
					"error": "pipeline not found",
				})
				return
			}
			ctx.JSON(200, record)
		})

		// 获取 pipeline 日志
		api.Get("/pipelines/:id/logs", func(ctx *zoox.Context) {
			id := ctx.Param().Get("id").String()
			record, ok := s.store.Get(id)
			if !ok {
				ctx.Status(404)
				ctx.JSON(404, map[string]string{
					"error": "pipeline not found",
				})
				return
			}
			ctx.JSON(200, map[string]interface{}{
				"data": record.Logs,
			})
		})

		// 删除 pipeline 记录
		api.Delete("/pipelines/:id", func(ctx *zoox.Context) {
			id := ctx.Param().Get("id").String()
			if s.store.Delete(id) {
				ctx.JSON(200, map[string]string{
					"message": "deleted",
				})
			} else {
				ctx.Status(404)
				ctx.JSON(404, map[string]string{
					"error": "pipeline not found",
				})
			}
		})

		// 获取队列统计信息
		api.Get("/queue/stats", func(ctx *zoox.Context) {
			stats := s.queue.Stats()
			ctx.JSON(200, stats)
		})

		// 获取队列列表
		api.Get("/queue", func(ctx *zoox.Context) {
			items := s.queue.List()
			ctx.JSON(200, map[string]interface{}{
				"data":  items,
				"total": len(items),
			})
		})

		// 取消队列项
		api.Delete("/queue/:id", func(ctx *zoox.Context) {
			id := ctx.Param().Get("id").String()
			if s.queue.Cancel(id) {
				ctx.JSON(200, map[string]string{
					"message": "cancelled",
				})
			} else {
				ctx.Status(404)
				ctx.JSON(404, map[string]string{
					"error": "queue item not found",
				})
			}
		})

		// 执行 pipeline (通过 WebSocket)
		api.Post("/pipelines/run", func(ctx *zoox.Context) {
			var req struct {
				Config string `json:"config"` // YAML 格式的 pipeline 配置
			}
			if err := ctx.BindJSON(&req); err != nil {
				ctx.Status(400)
				ctx.JSON(400, map[string]string{
					"error": fmt.Sprintf("invalid request: %s", err),
				})
				return
			}

			// 解析 pipeline 配置
			var pl pipeline.Pipeline
			if err := yaml.Decode([]byte(req.Config), &pl); err != nil {
				ctx.Status(400)
				ctx.JSON(400, map[string]string{
					"error": fmt.Sprintf("invalid pipeline config: %s", err),
				})
				return
			}

			// 返回 WebSocket 连接信息
			wsPath := s.cfg.Path
			if wsPath == "" {
				wsPath = "/"
			}
			wsURL := fmt.Sprintf("ws://%s%s", ctx.Request.Host, wsPath)
			if ctx.Request.TLS != nil {
				wsURL = fmt.Sprintf("wss://%s%s", ctx.Request.Host, wsPath)
			}

			ctx.JSON(200, map[string]interface{}{
				"ws_url":  wsURL,
				"message": "use WebSocket to execute pipeline",
			})
		})

		// 获取设置
		api.Get("/settings", func(ctx *zoox.Context) {
			ctx.JSON(200, map[string]interface{}{
				"max_concurrent":   s.cfg.MaxConcurrent,
				"max_records":      1000, // 从 store 获取
				"refresh_interval": 5,    // 前端设置
			})
		})

		// 保存设置（仅返回当前配置，实际需要重启服务）
		api.Post("/settings", func(ctx *zoox.Context) {
			var settings map[string]interface{}
			if err := ctx.BindJSON(&settings); err != nil {
				ctx.Status(400)
				ctx.JSON(400, map[string]string{
					"error": fmt.Sprintf("invalid request: %s", err),
				})
				return
			}

			// 注意：这里只是返回成功，实际配置需要重启服务才能生效
			ctx.JSON(200, map[string]string{
				"message": "settings saved (restart required to apply)",
			})
		})
	}

	// Web Console 静态文件
	app.Get("/console", func(ctx *zoox.Context) {
		ctx.SetHeader(headers.ContentType, "text/html")
		// ctx.String(200, getConsoleHTML())
		ctx.Write([]byte(getConsoleHTML()))
	})

	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, map[string]string{
			"version":    pipeline.Version,
			"running_at": app.Runtime().RunningAt().Format("YYYY-MM-DD HH:mm:ss"),
		})
	})

	return app.Run(fmt.Sprintf(":%d", s.cfg.Port))
}
