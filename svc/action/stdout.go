package action

import (
	"encoding/json"
)

const typeStdout = "stdout"

var Stdout = Create(
	typeStdout,
	func(log []byte) ([]byte, error) {
		act := Action{
			Type:    typeStdout,
			Payload: string(log),
		}

		return json.Marshal(act)
	},
	func(payload []byte) ([]byte, error) {
		return payload, nil
	},
)
