package main

import (
	"flag"
	"fmt"
	"os"

	"gqlsdl/schema"
)

func main() {
	url := flag.String("url", "", "GraphQL endpoint URL")
	output := flag.String("output", "schema.graphql", "Output file path")
	authHeader := flag.String("auth", "", "Authorization header value (optional)")
	referer := flag.String("referer", "", "Referer header value (optional)")
	origin := flag.String("origin", "", "Origin header value (optional)")

	flag.Parse()

	if *url == "" {
		fmt.Println("Usage: gqlsdl -url <graphql-endpoint> [-output <file>] [-auth <token>]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	opts := &schema.FetchOptions{
		Headers: make(map[string]string),
	}

	if *authHeader != "" {
		opts.Headers["Authorization"] = *authHeader
	}
	if *referer != "" {
		opts.Headers["Referer"] = *referer
	}
	if *origin != "" {
		opts.Headers["Origin"] = *origin
	}

	fmt.Printf("Fetching schema from %s...\n", *url)

	introspectionSchema, err := schema.FetchSchema(*url, opts)
	if err != nil {
		fmt.Printf("Error fetching schema: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Converting to SDL format...")
	sdl := schema.ConvertToSDL(introspectionSchema)

	fmt.Printf("Saving to %s...\n", *output)
	if err := schema.SaveToFile(sdl, *output); err != nil {
		fmt.Printf("Error saving file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done! Schema saved successfully.")
}
