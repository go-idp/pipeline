package pipeline

import (
	"context"
	"testing"

	"github.com/go-idp/pipeline/job"
	"github.com/go-idp/pipeline/stage"
	"github.com/go-idp/pipeline/step"
	"github.com/go-zoox/core-utils/fmt"
)

func TestPipeline(t *testing.T) {
	pipeline := &Pipeline{
		// Version: "0.0.0",
		Name: "XX 项目发布",
		// Image: "alpine:latest",
		Image: "whatwewant/zmicro:v1",
		Environment: map[string]string{
			"CI_GIT_SERVER":      "http://git.example.com",
			"CI_GIT_USERNAME":    "xxx",
			"CI_GIT_TOKEN":       "xxx",
			"CI_GIT_REPOSITORY":  "http://git.example.com/xxx/yyy.git",
			"CI_GIT_COMMIT_HASH": "64c13953ba1b1227f2dee9983d875ef85430c540",
		},
		Stages: []*stage.Stage{
			{
				Name: "checkout",
				Jobs: []*job.Job{
					{
						Name: "checkout source code",
						Steps: []*step.Step{
							{
								Name:    "检出",
								Command: `echo "checkout ($CI_GIT_REPOSITORY)"`,
							},
						},
					},
				},
			},
			{
				Name: "build",
				Jobs: []*job.Job{
					{
						Name: "build image",
						Steps: []*step.Step{
							{
								Name:    "install dependencies",
								Command: `echo "install dependencies done."`,
							},
							{
								Name:    "copy files",
								Command: `echo "copy files ..."`,
							},
							{
								Name:    "build from source code",
								Command: `echo "build from source code done."`,
							},
							{
								Name:    "push image to registry",
								Command: `echo "push image to registry done."`,
							},
						},
					},
				},
			},
			{
				Name: "deploy",
				Jobs: []*job.Job{
					{
						Name: "deploy to production",
						Steps: []*step.Step{
							{
								Name:    "deploy image in production",
								Command: `echo "deploy image in production done."`,
							},
						},
					},
				},
			},
		},
	}

	// fmt.PrintJSON(pipeline)

	if err := pipeline.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	fmt.PrintJSON(pipeline)
}
