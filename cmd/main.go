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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ostafen/suricata/pkg/gen"
	"github.com/ostafen/suricata/pkg/spec"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "suricata",
	}

	var genCmd = &cobra.Command{
		Use:          "gen [files...]",
		Short:        "Generate code from one or more spec YAML files",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE:         runGen,
	}

	rootCmd.AddCommand(genCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runGen(cmd *cobra.Command, args []string) error {
	var gen gen.CodeGenerator

	for _, specPath := range args {
		s, err := spec.LoadSpec(specPath)
		if err != nil {
			return err
		}

		code, err := gen.Generate(s)
		if err != nil {
			return err
		}

		path, name := splitPackage(s.Package)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(path, name)+".go", code, 0666); err != nil {
			return err
		}
	}
	return nil
}

func splitPackage(pkg string) (string, string) {
	parts := strings.Split(pkg, ".")
	return filepath.Join(parts[:]...), parts[len(parts)-1]
}
