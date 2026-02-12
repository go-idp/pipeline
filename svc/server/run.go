package server

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

			offset := 0
			if offsetStr := ctx.Request.URL.Query().Get("offset"); offsetStr != "" {
				if parsed, err := strconv.Atoi(offsetStr); err == nil {
					offset = parsed
				}
			}

			// 获取查询参数
			search := ctx.Request.URL.Query().Get("search")
			statusFilter := ctx.Request.URL.Query().Get("status")
			startTimeStr := ctx.Request.URL.Query().Get("start_time")
			endTimeStr := ctx.Request.URL.Query().Get("end_time")

			// 解析时间范围
			var startTime, endTime *time.Time
			if startTimeStr != "" {
				if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
					startTime = &t
				}
			}
			if endTimeStr != "" {
				if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
					endTime = &t
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

			// 应用过滤
			filtered := make([]*PipelineRecord, 0)
			for _, record := range records {
				// 搜索过滤
				if search != "" {
					searchLower := strings.ToLower(search)
					nameMatch := strings.Contains(strings.ToLower(record.Name), searchLower)
					idMatch := strings.Contains(strings.ToLower(record.ID), searchLower)
					if !nameMatch && !idMatch {
						continue
					}
				}

				// 状态过滤
				if statusFilter != "" && record.Status != statusFilter {
					continue
				}

				// 时间范围过滤
				if startTime != nil && record.StartedAt.Before(*startTime) {
					continue
				}
				if endTime != nil && record.StartedAt.After(*endTime) {
					continue
				}

				filtered = append(filtered, record)
			}

			// 应用分页
			total := len(filtered)
			if offset > 0 && offset < len(filtered) {
				filtered = filtered[offset:]
			}
			if limit > 0 && limit < len(filtered) {
				filtered = filtered[:limit]
			}

			ctx.JSON(200, map[string]interface{}{
				"data":   filtered,
				"total":  total,
				"limit":  limit,
				"offset": offset,
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

			logs := record.Logs

			// 应用过滤
			search := ctx.Request.URL.Query().Get("search")
			typeFilter := ctx.Request.URL.Query().Get("type")
			startTimeStr := ctx.Request.URL.Query().Get("start_time")
			endTimeStr := ctx.Request.URL.Query().Get("end_time")
			limit := 0
			if limitStr := ctx.Request.URL.Query().Get("limit"); limitStr != "" {
				if parsed, err := strconv.Atoi(limitStr); err == nil {
					limit = parsed
				}
			}
			offset := 0
			if offsetStr := ctx.Request.URL.Query().Get("offset"); offsetStr != "" {
				if parsed, err := strconv.Atoi(offsetStr); err == nil {
					offset = parsed
				}
			}

			// 解析时间范围
			var startTime, endTime *time.Time
			if startTimeStr != "" {
				if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
					startTime = &t
				}
			}
			if endTimeStr != "" {
				if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
					endTime = &t
				}
			}

			// 过滤日志
			filtered := make([]LogEntry, 0)
			for _, log := range logs {
				// 类型过滤
				if typeFilter != "" && log.Type != typeFilter {
					continue
				}

				// 搜索过滤
				if search != "" && !strings.Contains(strings.ToLower(log.Message), strings.ToLower(search)) {
					continue
				}

				// 时间范围过滤
				if startTime != nil && log.Timestamp.Before(*startTime) {
					continue
				}
				if endTime != nil && log.Timestamp.After(*endTime) {
					continue
				}

				filtered = append(filtered, log)
			}

			// 应用分页
			total := len(filtered)
			if offset > 0 && offset < len(filtered) {
				filtered = filtered[offset:]
			}
			if limit > 0 && limit < len(filtered) {
				filtered = filtered[:limit]
			}

			ctx.JSON(200, map[string]interface{}{
				"data":   filtered,
				"total":  total,
				"limit":  limit,
				"offset": offset,
			})
		})

		// 导出 pipeline 日志
		api.Get("/pipelines/:id/logs/export", func(ctx *zoox.Context) {
			id := ctx.Param().Get("id").String()
			record, ok := s.store.Get(id)
			if !ok {
				ctx.Status(404)
				ctx.JSON(404, map[string]string{
					"error": "pipeline not found",
				})
				return
			}

			format := ctx.Request.URL.Query().Get("format")
			if format == "" {
				format = "text"
			}

			logs := record.Logs

			// 应用过滤（与 logs 端点相同的过滤逻辑）
			search := ctx.Request.URL.Query().Get("search")
			typeFilter := ctx.Request.URL.Query().Get("type")
			startTimeStr := ctx.Request.URL.Query().Get("start_time")
			endTimeStr := ctx.Request.URL.Query().Get("end_time")

			var startTime, endTime *time.Time
			if startTimeStr != "" {
				if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
					startTime = &t
				}
			}
			if endTimeStr != "" {
				if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
					endTime = &t
				}
			}

			filtered := make([]LogEntry, 0)
			for _, log := range logs {
				if typeFilter != "" && log.Type != typeFilter {
					continue
				}
				if search != "" && !strings.Contains(strings.ToLower(log.Message), strings.ToLower(search)) {
					continue
				}
				if startTime != nil && log.Timestamp.Before(*startTime) {
					continue
				}
				if endTime != nil && log.Timestamp.After(*endTime) {
					continue
				}
				filtered = append(filtered, log)
			}

			if format == "json" {
				ctx.SetHeader(headers.ContentType, "application/json")
				ctx.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=pipeline-%s-logs.json", id))
				ctx.JSON(200, map[string]interface{}{
					"pipeline_id":   id,
					"pipeline_name": record.Name,
					"exported_at":   time.Now().Format(time.RFC3339),
					"logs":          filtered,
				})
			} else {
				// 文本格式
				ctx.SetHeader(headers.ContentType, "text/plain")
				ctx.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=pipeline-%s-logs.txt", id))

				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("Pipeline: %s (ID: %s)\n", record.Name, id))
				sb.WriteString(fmt.Sprintf("Exported at: %s\n", time.Now().Format(time.RFC3339)))
				sb.WriteString(strings.Repeat("=", 80) + "\n\n")

				for _, log := range filtered {
					timestamp := log.Timestamp.Format("2006-01-02 15:04:05")
					sb.WriteString(fmt.Sprintf("[%s] [%s] %s\n", timestamp, log.Type, log.Message))
				}

				ctx.String(200, sb.String())
			}
		})

		// 取消 pipeline 执行
		api.Post("/pipelines/:id/cancel", func(ctx *zoox.Context) {
			id := ctx.Param().Get("id").String()

			// 尝试从队列取消
			if s.queue != nil && s.queue.Cancel(id) {
				ctx.JSON(200, map[string]string{
					"message": "pipeline cancelled",
				})
				return
			}

			// 检查是否存在记录
			record, ok := s.store.Get(id)
			if !ok {
				ctx.Status(404)
				ctx.JSON(404, map[string]string{
					"error": "pipeline not found",
				})
				return
			}

			// 如果已经是最终状态，不能取消
			if record.Status == "succeeded" || record.Status == "failed" || record.Status == "cancelled" {
				ctx.Status(400)
				ctx.JSON(400, map[string]string{
					"error": fmt.Sprintf("pipeline is already %s, cannot cancel", record.Status),
				})
				return
			}

			// 如果不在队列中，直接更新状态为 cancelled
			if record.Status == "pending" || record.Status == "running" {
				s.store.UpdateStatus(id, "cancelled", fmt.Errorf("cancelled by user"))
				ctx.JSON(200, map[string]string{
					"message": "pipeline cancelled",
				})
				return
			}

			ctx.Status(400)
			ctx.JSON(400, map[string]string{
				"error": "pipeline cannot be cancelled",
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

		// 批量删除 pipeline 记录
		api.Post("/pipelines/batch/delete", func(ctx *zoox.Context) {
			var req struct {
				IDs []string `json:"ids"`
			}
			if err := ctx.BindJSON(&req); err != nil {
				ctx.Status(400)
				ctx.JSON(400, map[string]string{
					"error": fmt.Sprintf("invalid request: %s", err),
				})
				return
			}

			deleted := 0
			notFound := 0
			for _, id := range req.IDs {
				if s.store.Delete(id) {
					deleted++
				} else {
					notFound++
				}
			}

			ctx.JSON(200, map[string]interface{}{
				"message":   "batch delete completed",
				"deleted":   deleted,
				"not_found": notFound,
				"total":     len(req.IDs),
			})
		})

		// 批量取消 pipeline
		api.Post("/pipelines/batch/cancel", func(ctx *zoox.Context) {
			var req struct {
				IDs []string `json:"ids"`
			}
			if err := ctx.BindJSON(&req); err != nil {
				ctx.Status(400)
				ctx.JSON(400, map[string]string{
					"error": fmt.Sprintf("invalid request: %s", err),
				})
				return
			}

			cancelled := 0
			failed := 0
			notFound := 0

			for _, id := range req.IDs {
				// 尝试从队列取消
				if s.queue != nil && s.queue.Cancel(id) {
					cancelled++
					continue
				}

				// 检查是否存在记录
				record, ok := s.store.Get(id)
				if !ok {
					notFound++
					continue
				}

				// 如果已经是最终状态，不能取消
				if record.Status == "succeeded" || record.Status == "failed" || record.Status == "cancelled" {
					failed++
					continue
				}

				// 如果不在队列中，直接更新状态为 cancelled
				if record.Status == "pending" || record.Status == "running" {
					s.store.UpdateStatus(id, "cancelled", fmt.Errorf("cancelled by user"))
					cancelled++
				} else {
					failed++
				}
			}

			ctx.JSON(200, map[string]interface{}{
				"message":   "batch cancel completed",
				"cancelled": cancelled,
				"failed":    failed,
				"not_found": notFound,
				"total":     len(req.IDs),
			})
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

		// 配置转换 API
		// YAML 转可视化配置
		api.Post("/configs/convert/yaml-to-visual", func(ctx *zoox.Context) {
			var req struct {
				YAML string `json:"yaml"`
			}
			if err := ctx.BindJSON(&req); err != nil {
				ctx.Status(400)
				ctx.JSON(400, map[string]string{
					"error": fmt.Sprintf("invalid request: %s", err),
				})
				return
			}

			if s.configStore == nil {
				ctx.Status(500)
				ctx.JSON(500, map[string]string{
					"error": "config store not initialized",
				})
				return
			}

			visual, err := s.configStore.ConvertYAMLToVisual(req.YAML)
			if err != nil {
				ctx.Status(400)
				ctx.JSON(400, map[string]string{
					"error": fmt.Sprintf("failed to convert YAML to visual: %s", err),
				})
				return
			}

			ctx.JSON(200, map[string]interface{}{
				"visual": visual,
			})
		})

		// 可视化配置转 YAML
		api.Post("/configs/convert/visual-to-yaml", func(ctx *zoox.Context) {
			var req struct {
				Visual map[string]interface{} `json:"visual"`
			}
			if err := ctx.BindJSON(&req); err != nil {
				ctx.Status(400)
				ctx.JSON(400, map[string]string{
					"error": fmt.Sprintf("invalid request: %s", err),
				})
				return
			}

			if s.configStore == nil {
				ctx.Status(500)
				ctx.JSON(500, map[string]string{
					"error": "config store not initialized",
				})
				return
			}

			yaml, err := s.configStore.ConvertVisualToYAML(req.Visual)
			if err != nil {
				ctx.Status(400)
				ctx.JSON(400, map[string]string{
					"error": fmt.Sprintf("failed to convert visual to YAML: %s", err),
				})
				return
			}

			ctx.JSON(200, map[string]interface{}{
				"yaml": yaml,
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
