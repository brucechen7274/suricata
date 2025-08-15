// Copyright (c) 2025 Suricata Contributors
// Original Author: Stefano Scafiti
//
// This file is part of Suricata: Type-Safe AI Agents for Go.
//
// Licensed under the MIT License. You may obtain a copy of the License at
//
//	https://opensource.org/licenses/MIT
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"encoding/json"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

func UnmarshalValidate(data []byte, out any, schema gojsonschema.JSONLoader) error {
	if err := ValidateRawJSON(data, schema); err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func ValidateRawJSON(data []byte, schema gojsonschema.JSONLoader) error {
	res, err := gojsonschema.Validate(schema, gojsonschema.NewBytesLoader(data))
	if err != nil {
		return err
	}

	if !res.Valid() {
		return ErrInvalidOutput
	}
	return nil
}

func ValidateJSON(in any, schema gojsonschema.JSONLoader) error {
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return ValidateRawJSON(data, schema)
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
