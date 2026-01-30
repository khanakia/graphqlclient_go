package main

import (
	"fmt"
	"gqlsdk/generator"
)

func main() {
	config := &generator.Config{
		SchemaPath:  "../schema.graphql",
		OutputDir:   "../testsdk/api",
		PackageName: "testsdk/api",
		ModulePath:  "testsdk",
	}

	gen, err := generator.New(config)
	if err != nil {
		fmt.Println("failed to create generator: %w", err)
		return
	}

	if err := gen.Generate(); err != nil {
		fmt.Println("failed to generate SDK: %w", err)
		return
	}
}
