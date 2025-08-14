package runtime

import (
	"encoding/json"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

func UnmarshalValidate(data []byte, out any, schema gojsonschema.JSONLoader) error {
	err := json.Unmarshal(data, out)
	if err != nil {
		return err
	}

	res, err := gojsonschema.Validate(schema, gojsonschema.NewBytesLoader(data))
	if err != nil {
		return err
	}

	if !res.Valid() {
		return ErrInvalidOutput
	}
	return nil
}

// ExtractJSONFromString tries to find the first valid JSON object in the input string.
// It returns the JSON string and an error if none is found.
func ExtractJSONFromString(input string) string {
	start := strings.Index(input, "{")
	if start == -1 {
		return ""
	}

	braceCount := 0
	for i := start; i < len(input); i++ {
		switch input[i] {
		case '{':
			braceCount++
		case '}':
			braceCount--
			if braceCount == 0 {
				candidate := input[start : i+1]

				var js json.RawMessage
				if err := json.Unmarshal([]byte(candidate), &js); err == nil {
					return candidate
				}
			}
		}
	}
	return ""
}
