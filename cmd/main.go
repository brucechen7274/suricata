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

	if len(parts) == 1 {
		return ".", parts[0]
	}
	return filepath.Join(parts[:len(parts)-1]...), parts[len(parts)-1]
}
