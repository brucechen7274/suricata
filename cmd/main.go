package main

import (
	"fmt"
	"os"

	"github.com/ostafen/suricata/pkg/gen"
	"github.com/ostafen/suricata/pkg/spec"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: parser <spec.yaml>")
		os.Exit(1)
	}

	specPath := os.Args[1]
	spec, err := spec.LoadSpec(specPath)
	if err != nil {
		panic(err)
	}

	code, err := gen.Generate(spec)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(code))
}
