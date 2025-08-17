package graph

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

func MarshalJSON(v map[string]interface{}) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		b, err := json.Marshal(v)
		if err != nil {
			w.Write([]byte("{}"))
			return
		}
		w.Write(b)
	})
}

func UnmarshalJSON(v interface{}) (map[string]interface{}, error) {
	switch val := v.(type) {
	case map[string]interface{}:
		return val, nil
	case string:
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(val), &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON scalar: %w", err)
		}
		return m, nil
	default:
		return nil, fmt.Errorf("unexpected type %T for JSON scalar", v)
	}
}
