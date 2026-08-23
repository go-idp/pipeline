package server

import (
	"fmt"
	"time"

	"github.com/go-idp/pipeline"
	"github.com/go-idp/pipeline/job"
	"github.com/go-idp/pipeline/stage"
	"github.com/go-idp/pipeline/step"
	"github.com/go-zoox/encoding/yaml"
	"github.com/go-zoox/uuid"
	"github.com/go-zoox/zoox"
)

/* ---------- run definition (stage/job/step tree) ---------- */

// RunStepDef 是运行详情中的 step 节点。
type RunStepDef struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Command string `json:"command"`
	Status  string `json:"status"`
}

// RunJobDef 是运行详情中的 job 节点。
type RunJobDef struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Status string        `json:"status"`
	Steps  []*RunStepDef `json:"steps"`
}

// RunStageDef 是运行详情中的 stage 节点。
type RunStageDef struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Mode   string       `json:"mode"`
	Status string       `json:"status"`
	Jobs   []*RunJobDef `json:"jobs"`
}

// RunDefinition 是运行详情的执行结构树。
type RunDefinition struct {
	Stages []*RunStageDef `json:"stages"`
}

// RunDetail 是运行详情响应。
type RunDetail struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Status     string                 `json:"status"`
	Trigger    string                 `json:"trigger"`
	StartedAt  time.Time              `json:"started_at"`
	EndedAt    *time.Time             `json:"ended_at,omitempty"`
	Duration   int64                  `json:"duration"` // 秒
	Error      string                 `json:"error,omitempty"`
	YAML       string                 `json:"yaml,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
	Definition *RunDefinition         `json:"definition,omitempty"`
	States     map[string]*StageState `json:"states,omitempty"`
	Logs       []LogEntry             `json:"logs,omitempty"`
}

// buildRunDefinition 根据运行 ID 与原始 YAML 构建执行结构树，
// 复刻 prepare 阶段的 pre/post 注入与层级 ID 生成规则（stage = run.i,
// job = run.i.j, step = run.i.j.k），并与运行时状态快照合并。
func buildRunDefinition(runID, yamlStr string, states map[string]*StageState) *RunDefinition {
	def := &RunDefinition{Stages: make([]*RunStageDef, 0)}

	var pl pipeline.Pipeline
	if err := yaml.Decode([]byte(yamlStr), &pl); err != nil {
		return def
	}

	stages := pl.Stages

	// 复刻 prepare 中的 pre/post stage 注入
	if pl.Pre != "" {
		stages = append([]*stage.Stage{{
			Name: "pre",
			Jobs: []*job.Job{{
				Name: "pre",
				Steps: []*step.Step{{
					Name:    "pre",
					Command: pl.Pre,
				}},
			}},
		}}, stages...)
	}
	if pl.Post != "" {
		stages = append(stages, &stage.Stage{
			Name: "post",
			Jobs: []*job.Job{{
				Name: "post",
				Steps: []*step.Step{{
					Name:    "post",
					Command: pl.Post,
				}},
			}},
		})
	}

	for si, s := range stages {
		sid := fmt.Sprintf("%s.%d", runID, si)
		mode := s.RunMode
		if mode == "" {
			mode = "parallel"
		}

		stageDef := &RunStageDef{
			ID:     sid,
			Name:   s.Name,
			Mode:   mode,
			Status: nodeStatus(states, sid),
			Jobs:   make([]*RunJobDef, 0, len(s.Jobs)),
		}

		for ji, j := range s.Jobs {
			jid := fmt.Sprintf("%s.%d", sid, ji)
			jobDef := &RunJobDef{
				ID:     jid,
				Name:   j.Name,
				Status: nodeStatus(states, jid),
				Steps:  make([]*RunStepDef, 0, len(j.Steps)),
			}

			for ki, k := range j.Steps {
				kid := fmt.Sprintf("%s.%d", jid, ki)
				jobDef.Steps = append(jobDef.Steps, &RunStepDef{
					ID:      kid,
					Name:    k.Name,
					Command: k.Command,
					Status:  nodeStatus(states, kid),
				})
			}

			stageDef.Jobs = append(stageDef.Jobs, jobDef)
		}

		def.Stages = append(def.Stages, stageDef)
	}

	return def
}

// nodeStatus 返回节点状态，未开始（无状态快照）视为 pending。
func nodeStatus(states map[string]*StageState, id string) string {
	if states != nil {
		if state, ok := states[id]; ok && state.Status != "" {
			return state.Status
		}
	}

	return "pending"
}

// recordEndedAt 返回记录的结束时间。
func recordEndedAt(r *PipelineRecord) *time.Time {
	if r.SucceedAt != nil {
		return r.SucceedAt
	}
	if r.FailedAt != nil {
		return r.FailedAt
	}
	if r.CancelledAt != nil {
		return r.CancelledAt
	}
	return nil
}

// recordDuration 返回记录耗时（秒）。
func recordDuration(r *PipelineRecord) int64 {
	ended := recordEndedAt(r)
	if ended == nil {
		if r.Status == "running" {
			return int64(time.Since(r.StartedAt).Seconds())
		}
		return 0
	}
	d := ended.Sub(r.StartedAt).Seconds()
	if d < 0 {
		return 0
	}
	return int64(d)
}

// recordTrigger 返回记录的触发来源。
func recordTrigger(r *PipelineRecord) string {
	if r.Config != nil {
		if v, ok := r.Config["trigger"].(string); ok && v != "" {
			return v
		}
	}
	return "manual"
}

/* ---------- web api routes ---------- */

// MountWebAPI mounts the management-console REST API routes on the api group.
// It must be called after the base /api/v1 group is created.
func (s *server) MountWebAPI(api *zoox.RouterGroup) {
	// runs list (同 /pipelines，语义更贴合管理后台)
	api.Get("/runs", func(ctx *zoox.Context) {
		s.handleListPipelines(ctx)
	})

	// runs detail（含执行结构树与三级状态）
	api.Get("/runs/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id").String()
		record, ok := s.store.Get(id)
		if !ok {
			ctx.Status(404)
			ctx.JSON(404, map[string]string{"error": "run not found"})
			return
		}

		detail := &RunDetail{
			ID:         record.ID,
			Name:       record.Name,
			Status:     record.Status,
			Trigger:    recordTrigger(record),
			StartedAt:  record.StartedAt,
			EndedAt:    recordEndedAt(record),
			Duration:   recordDuration(record),
			Error:      record.Error,
			YAML:       record.YAML,
			Config:     record.Config,
			Definition: buildRunDefinition(record.ID, record.YAML, record.StagesState),
			States:     record.StagesState,
			Logs:       record.Logs,
		}

		ctx.JSON(200, detail)
	})

	// create run（从 YAML 入队）
	api.Post("/runs", func(ctx *zoox.Context) {
		var req struct {
			Config  string `json:"config"`
			Trigger string `json:"trigger"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			ctx.Status(400)
			ctx.JSON(400, map[string]string{"error": fmt.Sprintf("invalid request: %s", err)})
			return
		}

		var pl pipeline.Pipeline
		if err := yaml.Decode([]byte(req.Config), &pl); err != nil {
			ctx.Status(400)
			ctx.JSON(400, map[string]string{"error": fmt.Sprintf("invalid pipeline config: %s", err)})
			return
		}
		if pl.Name == "" {
			ctx.Status(400)
			ctx.JSON(400, map[string]string{"error": "pipeline name is required"})
			return
		}

		trigger := req.Trigger
		if trigger == "" {
			trigger = "manual"
		}

		id := uuid.V4()
		if err := s.queue.EnqueueWithYAMLAndTrigger(id, pl.Name, &pl, req.Config, trigger); err != nil {
			ctx.Status(500)
			ctx.JSON(500, map[string]string{"error": fmt.Sprintf("failed to enqueue pipeline: %s", err)})
			return
		}

		ctx.JSON(200, map[string]interface{}{
			"id":      id,
			"name":    pl.Name,
			"status":  "pending",
			"message": "enqueued",
		})
	})

	// rerun（复用原运行的定义）
	rerun := func(ctx *zoox.Context) {
		id := ctx.Param().Get("id").String()
		record, ok := s.store.Get(id)
		if !ok {
			ctx.Status(404)
			ctx.JSON(404, map[string]string{"error": "run not found"})
			return
		}
		if record.YAML == "" {
			ctx.Status(400)
			ctx.JSON(400, map[string]string{"error": "run has no yaml definition, cannot rerun"})
			return
		}

		var pl pipeline.Pipeline
		if err := yaml.Decode([]byte(record.YAML), &pl); err != nil {
			ctx.Status(400)
			ctx.JSON(400, map[string]string{"error": fmt.Sprintf("invalid pipeline config: %s", err)})
			return
		}

		newID := uuid.V4()
		if err := s.queue.EnqueueWithYAMLAndTrigger(newID, pl.Name, &pl, record.YAML, "rerun"); err != nil {
			ctx.Status(500)
			ctx.JSON(500, map[string]string{"error": fmt.Sprintf("failed to enqueue pipeline: %s", err)})
			return
		}

		ctx.JSON(200, map[string]interface{}{
			"id":      newID,
			"name":    pl.Name,
			"status":  "pending",
			"message": "rerun enqueued",
		})
	}
	api.Post("/runs/:id/rerun", rerun)
	api.Post("/pipelines/:id/rerun", rerun)

	// stages state（运行详情三级状态）
	api.Get("/pipelines/:id/stages", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id").String()
		record, ok := s.store.Get(id)
		if !ok {
			ctx.Status(404)
			ctx.JSON(404, map[string]string{"error": "run not found"})
			return
		}

		ctx.JSON(200, map[string]interface{}{
			"definition": buildRunDefinition(record.ID, record.YAML, record.StagesState),
			"states":     record.StagesState,
		})
	})

	/* ---------- pipeline config templates ---------- */

	// 内置模板/案例（只读，帮助快速上手）
	api.Get("/templates", func(ctx *zoox.Context) {
		ctx.JSON(200, map[string]interface{}{
			"data":  BuiltinTemplates,
			"total": len(BuiltinTemplates),
		})
	})

	api.Get("/configs", func(ctx *zoox.Context) {
		if s.configStore == nil {
			ctx.JSON(200, map[string]interface{}{"data": []interface{}{}, "total": 0})
			return
		}
		list := s.configStore.List()
		ctx.JSON(200, map[string]interface{}{"data": list, "total": len(list)})
	})

	api.Post("/configs", func(ctx *zoox.Context) {
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			YAML        string `json:"yaml"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			ctx.Status(400)
			ctx.JSON(400, map[string]string{"error": fmt.Sprintf("invalid request: %s", err)})
			return
		}

		tpl, err := s.configStore.Create(req.Name, req.Description, req.YAML)
		if err != nil {
			ctx.Status(400)
			ctx.JSON(400, map[string]string{"error": err.Error()})
			return
		}

		ctx.JSON(200, tpl)
	})

	api.Get("/configs/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id").String()
		tpl, ok := s.configStore.Get(id)
		if !ok {
			ctx.Status(404)
			ctx.JSON(404, map[string]string{"error": "config not found"})
			return
		}
		ctx.JSON(200, tpl)
	})

	api.Put("/configs/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id").String()
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			YAML        string `json:"yaml"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			ctx.Status(400)
			ctx.JSON(400, map[string]string{"error": fmt.Sprintf("invalid request: %s", err)})
			return
		}

		tpl, err := s.configStore.Update(id, req.Name, req.Description, req.YAML)
		if err != nil {
			ctx.Status(404)
			ctx.JSON(404, map[string]string{"error": err.Error()})
			return
		}
		ctx.JSON(200, tpl)
	})

	api.Delete("/configs/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id").String()
		if s.configStore.Delete(id) {
			ctx.JSON(200, map[string]string{"message": "deleted"})
		} else {
			ctx.Status(404)
			ctx.JSON(404, map[string]string{"error": "config not found"})
		}
	})
}
