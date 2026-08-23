package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-idp/pipeline"
	"github.com/go-zoox/encoding/yaml"
	"github.com/go-zoox/logger"
)

// seed 在服务启动时注入模拟数据：
//   - 配置模板库为空时注入内置模板（模板/案例，帮助用户快速上手）；
//   - 运行记录为空时注入演示运行（含日志与三级状态），让管理后台
//     首次打开即有内容可看。
//
// 已有数据（非空）不会被覆盖或重复注入。
func seed(configStore ConfigStore, store Store) {
	seedTemplates(configStore)
	seedDemoRuns(store)
}

// seedTemplates 注入内置模板（幂等：仅当模板库为空时）。
func seedTemplates(configStore ConfigStore) {
	if configStore == nil {
		return
	}

	if len(configStore.List()) > 0 {
		return
	}

	for _, tpl := range BuiltinTemplates {
		if _, err := configStore.Create(tpl.Name, tpl.Description, tpl.YAML); err != nil {
			logger.Warnf("[seed] failed to create template %s: %s", tpl.ID, err)
		}
	}

	logger.Infof("[seed] injected %d built-in templates", len(BuiltinTemplates))
}

// demoRun 描述一条演示运行记录。
type demoRun struct {
	tplID      string // 对应 BuiltinTemplates 的 ID
	name       string
	status     string // succeeded | failed
	failStage  int    // 失败所在 stage 下标（-1 表示全部成功）
	startedAgo time.Duration
}

var demoRuns = []*demoRun{
	{tplID: "ci", name: "CI", status: "succeeded", failStage: -1, startedAgo: 2 * time.Hour},
	{tplID: "deploy", name: "部署", status: "failed", failStage: 2, startedAgo: 26 * time.Hour},
	{tplID: "docs", name: "文档", status: "succeeded", failStage: -1, startedAgo: 50 * time.Hour},
}

// seedDemoRuns 注入演示运行记录（幂等：仅当运行记录为空时）。
func seedDemoRuns(store Store) {
	if store == nil {
		return
	}

	if len(store.List(1)) > 0 {
		return
	}

	for _, dr := range demoRuns {
		tpl := templateByID(dr.tplID)
		if tpl == nil {
			continue
		}

		id := fmt.Sprintf("demo-%s", dr.tplID)
		config := map[string]interface{}{
			"name":    dr.name,
			"workdir": fmt.Sprintf("/tmp/go-idp/pipeline/%s", id),
			"timeout": 3600,
			"trigger": "manual",
			"demo":    true,
		}

		store.CreateWithYAML(id, dr.name, tpl.YAML, config)

		// 回拨开始时间，让演示数据看起来是历史运行
		if rec, ok := store.Get(id); ok {
			rec.StartedAt = time.Now().Add(-dr.startedAgo)
		}

		seedDemoRunExecution(store, id, tpl, dr)

		var runErr error
		if dr.status == "failed" {
			runErr = fmt.Errorf("pipeline failed at stage %d", dr.failStage+1)
		}
		store.UpdateStatus(id, dr.status, runErr)
	}

	logger.Infof("[seed] injected %d demo runs", len(demoRuns))
}

func templateByID(id string) *Template {
	for _, tpl := range BuiltinTemplates {
		if tpl.ID == id {
			return tpl
		}
	}
	return nil
}

// seedDemoRunExecution 生成演示运行的日志与 stage/job/step 三级状态。
func seedDemoRunExecution(store Store, runID string, tpl *Template, dr *demoRun) {
	var pl pipeline.Pipeline
	if err := yaml.Decode([]byte(tpl.YAML), &pl); err != nil {
		return
	}

	start := time.Now().Add(-dr.startedAgo)
	var elapsed float64

	for si, s := range pl.Stages {
		sid := fmt.Sprintf("%s.%d", runID, si)
		stageFailed := dr.status == "failed" && si == dr.failStage
		stageStatus := "succeeded"
		if dr.status == "failed" && si > dr.failStage {
			stageStatus = "pending"
		} else if stageFailed {
			stageStatus = "failed"
		}

		for ji, j := range s.Jobs {
			jid := fmt.Sprintf("%s.%d", sid, ji)
			for ki, k := range j.Steps {
				kid := fmt.Sprintf("%s.%d", jid, ki)
				stepFailed := stageFailed && ki == 0
				stepStatus := "succeeded"
				if dr.status == "failed" && (si > dr.failStage || (si == dr.failStage && ki > 0)) {
					stepStatus = "pending"
				} else if stepFailed {
					stepStatus = "failed"
				}

				// 演示日志
				lines := demoStepLogs(k.Name, k.Command, stepStatus)
				for _, ln := range lines {
					elapsed += 0.4
					store.AddLog(runID, ln.typ, ln.text, start.Add(time.Duration(elapsed)*time.Second))
				}

				// 状态快照
				store.UpsertStageState(runID, &StageState{
					ID:        kid,
					Level:     "step",
					Name:      k.Name,
					Status:    stepStatus,
					StartedAt: start.Add(time.Duration(elapsed) * time.Second),
				})
			}

			store.UpsertStageState(runID, &StageState{
				ID:        jid,
				Level:     "job",
				Name:      j.Name,
				Status:    stageStatus,
				StartedAt: start.Add(time.Duration(elapsed) * time.Second),
			})
		}

		store.UpsertStageState(runID, &StageState{
			ID:        sid,
			Level:     "stage",
			Name:      s.Name,
			Status:    stageStatus,
			StartedAt: start.Add(time.Duration(elapsed) * time.Second),
		})
	}
}

type demoLogLine struct {
	typ  string // stdout | stderr
	text string
}

// demoStepLogs 根据步骤名生成逼真的演示日志。
func demoStepLogs(stepName, command, status string) []demoLogLine {
	out := []demoLogLine{{typ: "stdout", text: "$ " + command}}
	k := strings.ToLower(stepName)

	switch {
	case strings.Contains(k, "checkout") || strings.Contains(k, "clone"):
		out = append(out,
			demoLogLine{"stdout", "Cloning into '.'..."},
			demoLogLine{"stdout", "remote: Enumerating objects: 320, done."},
			demoLogLine{"stdout", "Receiving objects: 100% (320/320), 18.42 MiB | 9.8 MiB/s, done."},
			demoLogLine{"stdout", "Resolving deltas: 100% (150/150), done."},
		)
	case strings.Contains(k, "build") || strings.Contains(k, "bench"):
		out = append(out,
			demoLogLine{"stdout", "go: downloading github.com/go-zoox/zoox v1.18.14"},
			demoLogLine{"stdout", "go: downloading github.com/go-zoox/fs v1.4.1"},
			demoLogLine{"stdout", "ok  	github.com/go-idp/pipeline	0.462s"},
		)
	case strings.Contains(k, "test") || strings.Contains(k, "lint"):
		out = append(out,
			demoLogLine{"stdout", "ok  	github.com/go-idp/pipeline/job	0.118s"},
			demoLogLine{"stdout", "ok  	github.com/go-idp/pipeline/stage	0.203s"},
			demoLogLine{"stdout", "ok  	github.com/go-idp/pipeline/step	0.311s"},
		)
	case strings.Contains(k, "docker") || strings.Contains(k, "push") || strings.Contains(k, "deploy"):
		out = append(out,
			demoLogLine{"stdout", "#1 [internal] load build definition from Dockerfile"},
			demoLogLine{"stdout", "#4 [1/5] FROM docker.io/library/golang:1.21-alpine@sha256:abc123"},
			demoLogLine{"stdout", "#8 [5/5] RUN go build -o /app/pipeline ./cmd/pipeline"},
			demoLogLine{"stdout", "#9 exporting layers 18.6s done"},
			demoLogLine{"stdout", "#9 pushing manifest for registry.example.com/idp/backend:test-pipeline"},
		)
	case strings.Contains(k, "docs") || strings.Contains(k, "pages"):
		out = append(out,
			demoLogLine{"stdout", "vitepress v1.6.4"},
			demoLogLine{"stdout", "building client + server bundles..."},
			demoLogLine{"stdout", "rendering pages..."},
			demoLogLine{"stdout", "build complete in 6.2s."},
		)
	default:
		out = append(out,
			demoLogLine{"stdout", "[INFO] start"},
			demoLogLine{"stdout", "[INFO] processing..."},
		)
	}

	if status == "failed" {
		out = append(out,
			demoLogLine{"stderr", "--- FAIL: step " + stepName},
			demoLogLine{"stderr", "FAIL"},
			demoLogLine{"stderr", "✘ Step \"" + stepName + "\" failed"},
		)
	} else {
		out = append(out, demoLogLine{"stdout", "✔ Step \"" + stepName + "\" succeeded"})
	}

	return out
}
