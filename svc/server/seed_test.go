package server

import (
	"testing"
	"time"

	"github.com/go-idp/pipeline"
	"github.com/go-zoox/encoding/yaml"
)

// TestBuiltinTemplatesValid 验证所有内置模板都是合法可解析的流水线定义。
func TestBuiltinTemplatesValid(t *testing.T) {
	if len(BuiltinTemplates) == 0 {
		t.Fatal("expected built-in templates, got none")
	}

	for _, tpl := range BuiltinTemplates {
		t.Run(tpl.ID, func(t *testing.T) {
			if tpl.Name == "" {
				t.Fatalf("template %s has no name", tpl.ID)
			}
			if tpl.YAML == "" {
				t.Fatalf("template %s has no yaml", tpl.ID)
			}

			var pl pipeline.Pipeline
			if err := yaml.Decode([]byte(tpl.YAML), &pl); err != nil {
				t.Fatalf("template %s yaml is invalid: %v", tpl.ID, err)
			}
			if pl.Name == "" {
				t.Fatalf("template %s pipeline name is empty", tpl.ID)
			}
			if len(pl.Stages) == 0 {
				t.Fatalf("template %s has no stages", tpl.ID)
			}
		})
	}
}

// TestSeedTemplatesIdempotent 验证模板注入的幂等性：空库注入，非空不覆盖。
func TestSeedTemplatesIdempotent(t *testing.T) {
	store := NewMemoryConfigStore("")

	// 空库 -> 注入全部内置模板
	seedTemplates(store)
	if got := len(store.List()); got != len(BuiltinTemplates) {
		t.Fatalf("expected %d templates after seed, got %d", len(BuiltinTemplates), got)
	}

	// 再次注入 -> 数量不变
	seedTemplates(store)
	if got := len(store.List()); got != len(BuiltinTemplates) {
		t.Fatalf("expected %d templates after second seed, got %d", len(BuiltinTemplates), got)
	}

	// 非空库（用户自定义）-> 不注入
	userStore := NewMemoryConfigStore("")
	if _, err := userStore.Create("My Own", "", "name: my\nstages:\n  - name: a\n    jobs:\n      - name: j\n        steps:\n          - name: s\n            command: echo hi"); err != nil {
		t.Fatalf("failed to create user template: %v", err)
	}
	seedTemplates(userStore)
	if got := len(userStore.List()); got != 1 {
		t.Fatalf("expected 1 user template preserved, got %d", got)
	}
}

// TestSeedDemoRunsIdempotent 验证演示运行注入的幂等性。
func TestSeedDemoRunsIdempotent(t *testing.T) {
	store := NewMemoryStore("", 100)

	seedDemoRuns(store)
	if got := len(store.List(100)); got != len(demoRuns) {
		t.Fatalf("expected %d demo runs after seed, got %d", len(demoRuns), got)
	}

	// 再次注入 -> 数量不变
	seedDemoRuns(store)
	if got := len(store.List(100)); got != len(demoRuns) {
		t.Fatalf("expected %d demo runs after second seed, got %d", len(demoRuns), got)
	}
}

// TestSeedDemoRunContent 验证演示运行包含日志、三级状态与回拨的开始时间。
func TestSeedDemoRunContent(t *testing.T) {
	store := NewMemoryStore("", 100)
	seedDemoRuns(store)

	rec, ok := store.Get("demo-ci")
	if !ok {
		t.Fatal("demo-ci run not found")
	}
	if rec.Status != "succeeded" {
		t.Fatalf("demo-ci status: got %q, want succeeded", rec.Status)
	}

	// 开始时间已回拨（早于现在 1 小时以上）
	if !rec.StartedAt.Before(time.Now().Add(-time.Hour)) {
		t.Fatalf("demo-ci started_at should be backdated, got %v", rec.StartedAt)
	}

	// 日志非空
	if len(rec.Logs) == 0 {
		t.Fatal("demo-ci has no logs")
	}

	// 三级状态：stage + job + step 快照均已注入
	if len(rec.StagesState) < 3 {
		t.Fatalf("demo-ci stages_state: expected >=3 nodes, got %d", len(rec.StagesState))
	}
	hasStep := false
	hasStage := false
	for _, st := range rec.StagesState {
		if st.Level == "step" {
			hasStep = true
		}
		if st.Level == "stage" {
			hasStage = true
		}
	}
	if !hasStep || !hasStage {
		t.Fatalf("demo-ci stages_state should contain both stage and step nodes")
	}

	// 失败演示运行应包含失败状态与 stderr 日志
	failed, ok := store.Get("demo-deploy")
	if !ok {
		t.Fatal("demo-deploy run not found")
	}
	if failed.Status != "failed" {
		t.Fatalf("demo-deploy status: got %q, want failed", failed.Status)
	}
	hasStderr := false
	for _, l := range failed.Logs {
		if l.Type == "stderr" {
			hasStderr = true
			break
		}
	}
	if !hasStderr {
		t.Fatal("demo-deploy should contain stderr log lines")
	}
}
