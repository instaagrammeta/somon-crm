package handler

import "encoding/json"

// jsonMarshal is a small wrapper to avoid importing encoding/json everywhere.
func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
