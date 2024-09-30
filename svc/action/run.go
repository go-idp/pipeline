package action

import (
	"encoding/json"
	"fmt"

	"github.com/go-idp/pipeline"
	"github.com/go-zoox/encoding/yaml"
)

const typeRun = "run"

var Run = Create(
	typeRun,
	func(pl *pipeline.Pipeline) ([]byte, error) {
		payload, err := yaml.Encode(pl)
		if err != nil {
			return nil, fmt.Errorf("failed to encode run action: %s", err)
		}

		act := Action{
			Type:    typeRun,
			Payload: string(payload),
		}

		return json.Marshal(act)
	},
	func(payload []byte) (*pipeline.Pipeline, error) {
		pl := pipeline.Pipeline{}
		if err := yaml.Decode(payload, &pl); err != nil {
			return nil, fmt.Errorf("failed to decode run action: %s", err)
		}

		return &pl, nil
	},
)
