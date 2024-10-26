package action

import (
	"encoding/json"
)

const typeStderr = "stderr"

var Stderr = Create(
	typeStderr,
	func(log []byte) ([]byte, error) {
		act := Action{
			Type:    typeStderr,
			Payload: string(log),
		}

		return json.Marshal(act)
	},
	func(payload []byte) ([]byte, error) {
		return payload, nil
	},
)
