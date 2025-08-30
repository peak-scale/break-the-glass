package items

import (
	json2 "encoding/json"
)

func (ps ParamSchema) JSON() []byte {
	json, _ := json2.Marshal(ps.Object)
	return json
}
