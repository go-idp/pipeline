package action

import (
	"encoding/json"
	"fmt"
)

const typeDone = "done"

var Done = Create(
	typeDone,
	func(err error) ([]byte, error) {
		act := Action{
			Type: typeDone,
		}

		if err != nil {
			act.Payload = err.Error()
		}

		return json.Marshal(act)
	},
	func(payload []byte) (error, error) {
		if len(payload) == 0 {
			return nil, nil
		}

		return fmt.Errorf("%s", string(payload)), nil
	},
)
