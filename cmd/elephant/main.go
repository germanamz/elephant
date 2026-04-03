package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "elephant",
		Short:   "AI orchestration platform",
		Version: fmt.Sprintf("%s (%s, %s)", version, commit, date),
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
