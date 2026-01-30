package main

import (
	"fmt"
	"gqlcustom/pkg/clientgen"
)

func main() {
	fmt.Println("Hello, World!")

	config := &clientgen.Config{
		SchemaPath: "cmd/test/schema.graphql",
		// SchemaPath:  "cmd/test/enum.graphql",
		OutputDir:   "./sdk",
		PackageName: "sdk",
		ModulePath:  "github.com/example/sdk",
		ConfigPath:  "cmd/test/config.jsonc",
		Package:     "gqlcustom/sdk",
	}

	gen, err := clientgen.New(config)
	if err != nil {
		fmt.Println("failed to create generator: %w", err)
		return
	}

	if err := gen.Generate(); err != nil {
		fmt.Println("failed to generate SDK: %w", err)
		return
	}
}
