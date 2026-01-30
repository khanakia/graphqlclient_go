package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gqlsdk",
	Short: "GraphQL SDK Generator",
	Long:  `A CLI tool that generates type-safe Go client SDKs from GraphQL SDL files.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
