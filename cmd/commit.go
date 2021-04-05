package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commits a dependency structure.",
	Long:  `Creates a tree of dependencies and attempts to commit them.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.Unmarshal(&cfg)
		fmt.Printf("Command currently not supported.")
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}
