package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-idp/pipeline/stage"
	"github.com/go-zoox/encoding/yaml"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/uuid"
)

type Pipeline struct {
	Name   string         `json:"name" yaml:"name"`
	Stages []*stage.Stage `json:"stages" yaml:"stages"`
	//
	Workdir string `json:"workdir" yaml:"workdir"`
	//
	Image       string            `json:"image" yaml:"image"`
	Environment map[string]string `json:"environment" yaml:"environment"`
	//
	State *State `json:"state" yaml:"state"`
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

	if p.Workdir == "" {
		p.Workdir = fs.CurrentDir()
	}

	if ok := fs.IsExist(p.Workdir); !ok {
		logger.Infof("[workflow][prepare] create workdir(path: %s)", p.Workdir)
		if err := fs.Mkdirp(p.Workdir); err != nil {
			return fmt.Errorf("[workflow][prepare] failed to create workdir(path: %s): %s", p.Workdir, err)
		}
	} else {
		if ok := fs.IsEmpty(p.Workdir); !ok {
			return fmt.Errorf("[workflow][prepare] workdir(path: %s) is not empty", p.Workdir)
		}
	}

	if p.Environment == nil {
		p.Environment = make(map[string]string)
	}
	p.Environment["PIPELINE_VERSION"] = Version
	p.Environment["PIPELINE_NAME"] = p.Name
	p.Environment["PIPELINE_WORKDIR"] = p.Workdir

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
			Image:       p.Image,
			Environment: p.Environment,
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
	defer logger.Infof("[workflow] done to run (name: %s)", p.Name)

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

func (p *Pipeline) SetWorkdir(workdir string) {
	p.Workdir = workdir
}

func (p *Pipeline) SetEnvironment(environment map[string]string) {
	if p.Environment == nil {
		p.Environment = make(map[string]string)
	}

	for k, v := range environment {
		if _, ok := p.Environment[k]; !ok {
			p.Environment[k] = v
		}
	}
}

func (p *Pipeline) SetStdout(stdout io.Writer) {
	p.stdout = stdout

	for _, stage := range p.Stages {
		stage.SetStdout(stdout)
	}
}

func (p *Pipeline) SetStderr(stderr io.Writer) {
	p.stderr = stderr

	for _, stage := range p.Stages {
		stage.SetStderr(stderr)
	}
}
