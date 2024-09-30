package action

import (
	"encoding/json"
)

const typeLog = "log"

var Log = Create(
	typeLog,
	func(log []byte) ([]byte, error) {
		act := Action{
			Type:    typeLog,
			Payload: string(log),
		}

		return json.Marshal(act)
	},
	func(payload []byte) ([]byte, error) {
		return payload, nil
	},
)
