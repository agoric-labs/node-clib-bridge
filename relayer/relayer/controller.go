package relayer

import (
	"encoding/json"
)

var SendToController func(needReply bool, str string) (string, error)

func ControllerUpcall(action interface{}) (bool, error) {
	bz, err := json.Marshal(action)
	if err != nil {
		return false, err
	}
	ret, err := SendToController(true, string(bz))
	if err != nil {
		return false, err
	}
	return ret == "true", nil
}
