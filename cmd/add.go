package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a project to your dependency.",
	Long:  `Add a project to your dependency.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		viper.Unmarshal(&cfg)
		for _, arg := range args {
			cfg.Projects = appendIfMissing(cfg.Projects, arg)
		}
		viper.Set("Projects", cfg.Projects)
		viper.WriteConfig()
		fmt.Println(cfg.String())
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func appendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}
