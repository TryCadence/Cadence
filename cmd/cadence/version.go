package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codemeapixel/cadence/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display the version of Cadence and build information`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Full())
	},
}
