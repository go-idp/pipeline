package job

import (
	"testing"

	"github.com/go-idp/pipeline/step"
)

func TestJobSetup_MergeAndPropagateToSteps(t *testing.T) {
	j := &Job{
		Name: "job",
		Steps: []*step.Step{
			{Name: "s1"},
			// should not be overridden by job.Workdir
			{Name: "s2", Workdir: "/custom"},
		},
	}

	opt := &Job{
		Image:   "alpine:3",
		Workdir: "/work",
		Environment: map[string]string{
			"FOO": "bar",
		},
		Timeout: 123,
	}

	if err := j.Setup("jid", opt); err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	if j.State == nil {
		t.Fatalf("job state is nil")
	}
	if j.State.ID != "jid" {
		t.Fatalf("job state id mismatch: got %q", j.State.ID)
	}
	if j.State.Status != "running" {
		t.Fatalf("job state status mismatch: got %q", j.State.Status)
	}
	if j.State.StartedAt.IsZero() {
		t.Fatalf("job started_at is zero")
	}

	if j.Image != opt.Image {
		t.Fatalf("job image mismatch: got %q want %q", j.Image, opt.Image)
	}
	if j.Workdir != opt.Workdir {
		t.Fatalf("job workdir mismatch: got %q want %q", j.Workdir, opt.Workdir)
	}
	if j.Timeout != opt.Timeout {
		t.Fatalf("job timeout mismatch: got %d want %d", j.Timeout, opt.Timeout)
	}
	if j.Environment == nil || j.Environment["FOO"] != "bar" {
		t.Fatalf("job environment not merged/propagated: %#v", j.Environment)
	}

	// step 1 inherits job config
	s1 := j.Steps[0]
	if s1.State == nil || s1.State.ID != "jid.0" {
		t.Fatalf("step1 state id mismatch: %#v", s1.State)
	}
	if s1.Workdir != "/work" {
		t.Fatalf("step1 workdir mismatch: got %q want %q", s1.Workdir, "/work")
	}
	if s1.Image != "alpine:3" {
		t.Fatalf("step1 image mismatch: got %q want %q", s1.Image, "alpine:3")
	}
	if s1.Timeout != 123 {
		t.Fatalf("step1 timeout mismatch: got %d want %d", s1.Timeout, 123)
	}
	if s1.Environment == nil || s1.Environment["FOO"] != "bar" {
		t.Fatalf("step1 environment mismatch: %#v", s1.Environment)
	}

	// step 2 should keep its own workdir
	s2 := j.Steps[1]
	if s2.Workdir != "/custom" {
		t.Fatalf("step2 workdir should not be overridden: got %q want %q", s2.Workdir, "/custom")
	}
}

