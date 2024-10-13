package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/go-idp/pipeline/job"
	"github.com/go-idp/pipeline/stage"
	"github.com/go-idp/pipeline/step"
	"github.com/go-zoox/encoding/yaml"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/safe"
	"github.com/go-zoox/uuid"
)

type Pipeline struct {
	Name string `json:"name" yaml:"name"`
	//
	Stages []*stage.Stage `json:"stages" yaml:"stages"`
	//
	Workdir string `json:"workdir" yaml:"workdir"`
	//
	Environment map[string]string `json:"environment" yaml:"environment"`
	//
	Image string `json:"image" yaml:"image"`
	//
	State *State `json:"state" yaml:"state"`
	//
	Pre  string `json:"pre" yaml:"pre"`
	Post string `json:"post" yaml:"post"`
	//
	stdout io.Writer
	stderr io.Writer
}

func (p *Pipeline) getLogger() *logger.Logger {
	l := logger.New()
	l.SetStdout(p.stdout)
	return l
}

func (p *Pipeline) prepare(id string) error {
	if p.stdout == nil {
		p.stdout = os.Stdout
	}

	if p.stderr == nil {
		p.stderr = p.stdout
	}

	logger := p.getLogger()
	logger.Infof("[workflow][prepare] start ...")
	defer logger.Infof("[workflow][prepare] done")

	if p.Name == "" {
		return fmt.Errorf("[workflow][prepare] name is required")
	}

	if p.Workdir == "" {
		p.Workdir = fs.CurrentDir()
	}

	// if workdir is current dir, skip create
	if p.Workdir != fs.CurrentDir() {
		if ok := fs.IsExist(p.Workdir); !ok {
			logger.Infof("[workflow][prepare] create workdir(path: %s)", p.Workdir)
			if err := fs.Mkdirp(p.Workdir); err != nil {
				return fmt.Errorf("[workflow][prepare] failed to create workdir(path: %s): %s", p.Workdir, err)
			}
		}
	}

	if p.Environment == nil {
		p.Environment = make(map[string]string)
	} else {
		// avoid nested pipeline
		if _, ok := p.Environment["PIPELINE_RUNNER"]; ok {
			return fmt.Errorf("[workflow][prepare] you are already in a pipeline, nested pipeline is not allowed")
		}
	}
	p.Environment["PIPELINE_RUNNER"] = "pipeline"
	p.Environment["PIPELINE_RUNNER_OS"] = runtime.GOOS
	p.Environment["PIPELINE_RUNNER_ARCH"] = runtime.GOARCH
	p.Environment["PIPELINE_RUNNER_VERSION"] = Version
	p.Environment["PIPELINE_RUNNER_USER"] = os.Getenv("USER")
	p.Environment["PIPELINE_RUNNER_WORKDIR"] = fs.CurrentDir()
	//
	p.Environment["PIPELINE_NAME"] = p.Name
	p.Environment["PIPELINE_WORKDIR"] = p.Workdir

	if len(p.Stages) == 0 {
		return fmt.Errorf("[workflow][prepare] no stages found, stages is required")
	}

	// add pre/post stage
	if p.Pre != "" {
		p.Stages = append([]*stage.Stage{
			{
				Name: "pre",
				Jobs: []*job.Job{
					{
						Name: "pre",
						Steps: []*step.Step{
							{
								Name:    "pre",
								Command: p.Pre,
							},
						},
					},
				},
			},
		}, p.Stages...)
	}
	if p.Post != "" {
		p.Stages = append(p.Stages, &stage.Stage{
			Name: "post",
			Jobs: []*job.Job{
				{
					Name: "post",
					Steps: []*step.Step{
						{
							Name:    "post",
							Command: p.Post,
						},
					},
				},
			},
		})
	}

	// setup state
	p.State = &State{
		ID:     id,
		Status: "running",
		//
		StartedAt: time.Now(),
	}

	for index, s := range p.Stages {
		err := s.Setup(fmt.Sprintf("%s.%d", p.State.ID, index), &stage.Stage{
			Workdir: p.Workdir,
			//
			Environment: p.Environment,
			//
			Image: p.Image,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pipeline) clean() error {
	logger := p.getLogger()
	logger.Infof("[workflow][clean] start ...")
	defer logger.Infof("[workflow][clean] done")

	if p.Workdir == "" {
		return nil
	}

	// fix: if workdir is removed, fs.CurrentDir() panic => cannot get current dir with os.Getwd(): getwd: no such file or directory
	err := safe.Do(func() error {
		fs.CurrentDir()
		return nil
	})
	if err != nil {
		return nil
	}

	// if workdir is current dir, skip clean
	if p.Workdir == fs.CurrentDir() {
		return nil
	}

	if ok := fs.IsExist(p.Workdir); !ok {
		return nil
	}

	logger.Infof("[workflow][clean] clean workdir(path: %s)", p.Workdir)
	if err := fs.RemoveDir(p.Workdir); err != nil {
		return fmt.Errorf("[workflow][clean] failed to clean workdir(path: %s): %s", p.Workdir, err)
	}

	return nil
}

func (p *Pipeline) Run(ctx context.Context, id ...string) error {
	//
	logger.Infof("[workflow] start to run (name: %s)", p.Name)
	defer logger.Infof("[workflow] done to run (name: %s, workdir: %s)", p.Name, p.Workdir)

	_id := uuid.V4()
	if len(id) > 0 {
		_id = id[0]
	}
	if err := p.prepare(_id); err != nil {
		return err
	}
	defer p.clean()

	plog := p.getLogger()
	plog.Infof("[workflow] start")
	plog.Infof("[workflow] version: %s", Version)
	plog.Infof("[workflow] name: %s", p.Name)
	plog.Infof("[workflow] workdir: %s", p.Workdir)
	defer plog.Infof("[workflow] done")

	for _, stage := range p.Stages {
		if err := stage.Run(ctx); err != nil {
			p.State.Status = "failed"
			p.State.Error = err.Error()
			p.State.FailedAt = time.Now()
			return err
		}
	}

	p.State.Status = "succeeded"
	p.State.SucceedAt = time.Now()

	return nil
}

func (p *Pipeline) String() string {
	v, err := yaml.Encode(p)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return string(v)
}

func (p *Pipeline) SetWorkdir(workdir string) *Pipeline {
	p.Workdir = workdir
	return p
}

func (p *Pipeline) SetEnvironment(environment map[string]string) *Pipeline {
	if p.Environment == nil {
		p.Environment = make(map[string]string)
	}

	for k, v := range environment {
		if _, ok := p.Environment[k]; !ok {
			p.Environment[k] = v
		}
	}

	return p
}

func (p *Pipeline) SetImage(image string) *Pipeline {
	p.Image = image
	return p
}

func (p *Pipeline) SetStdout(stdout io.Writer) *Pipeline {
	p.stdout = stdout

	for _, stage := range p.Stages {
		stage.SetStdout(stdout)
	}

	return p
}

func (p *Pipeline) SetStderr(stderr io.Writer) *Pipeline {
	p.stderr = stderr

	for _, stage := range p.Stages {
		stage.SetStderr(stderr)
	}

	return p
}
