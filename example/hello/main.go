// Copyright (c) 2025 Suricata Contributors
// Original Author: Stefano Scafiti
//
// This file is part of Suricata: Type-Safe AI Agents for Go.
//
// Licensed under the MIT License. You may obtain a copy of the License at
//
//     https://opensource.org/licenses/MIT
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"

	"github.com/ostafen/suricata/example/hello/hello"
	"github.com/ostafen/suricata/runtime/ollama"
)

func main() {
	invoker := ollama.NewInvoker(
		ollama.DefaultBaseURL,
		"granite3.3:8b",
		ollama.Options{
			NumCtx:      131072,
			Temperature: 0.1,
		},
	)

	helloAgent := hello.NewHelloAgent(invoker, &tools{})

	res, err := helloAgent.SayHelloAll(context.Background(), &hello.SayHelloAllRequest{
		Names: []string{"Pippo", "Pluto"},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(res.Ok)
}

type tools struct{}

func (t *tools) SayHelloTool(ctx context.Context, in *hello.SayHelloToolRequest) (*hello.SayHelloToolReply, error) {
	fmt.Println("Hello " + in.Name)

	return &hello.SayHelloToolReply{Ok: true}, nil
}
