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

	"github.com/ostafen/suricata/example/calculator/eval"
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

	cli := eval.NewMathAgent(invoker, &mathTools{})

	expr := "(8 + 3) * (2 - 4 / 2)"

	fmt.Printf("Evaluating expression: %s\n", expr)

	res, err := cli.Evaluate(context.Background(), &eval.EvalRequest{
		Expr: expr,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Result: %f!\n", res.Result)
}

type mathTools struct{}

func (t *mathTools) AddTool(ctx context.Context, in *eval.MathRequest) (*eval.MathReply, error) {
	fmt.Printf("%f + %f = %f\n", in.A, in.B, in.A+in.B)

	return &eval.MathReply{
		Result: in.A + in.B,
	}, nil
}

func (t *mathTools) SubTool(ctx context.Context, in *eval.MathRequest) (*eval.MathReply, error) {
	fmt.Printf("%f - %f = %f\n", in.A, in.B, in.A-in.B)

	return &eval.MathReply{
		Result: in.A - in.B,
	}, nil
}

func (t *mathTools) MulTool(ctx context.Context, in *eval.MathRequest) (*eval.MathReply, error) {
	fmt.Printf("%f * %f = %f\n", in.A, in.B, in.A*in.B)

	return &eval.MathReply{
		Result: in.A * in.B,
	}, nil
}

func (t *mathTools) DivTool(ctx context.Context, in *eval.MathRequest) (*eval.MathReply, error) {
	if in.B == 0 {
		fmt.Printf("%f / %f: division by zero!\n", in.A, in.B)

		return nil, fmt.Errorf("division by zero")
	}

	fmt.Printf("%f / %f = %f\n", in.A, in.B, in.A/in.B)

	return &eval.MathReply{
		Result: in.A / in.B,
	}, nil
}
