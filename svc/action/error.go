package action

import (
	"encoding/json"
	"fmt"
)

const typeError = "error"

var Error = Create(
	typeError,
	func(err error) ([]byte, error) {
		act := Action{
			Type:    typeError,
			Payload: err.Error(),
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
