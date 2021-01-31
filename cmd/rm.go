package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove a project from your dependency.",
	Long:  `Remove a project from your dependency.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.Unmarshal(&cfg)
		pros := make([]string, len(cfg.Projects))
		for _, arg := range args {
			for _, ele := range cfg.Projects {
				if arg != ele {
					pros = append(pros, ele)
				}
			}
		}
		cfg.Projects = pros
		viper.Set("Projects", cfg.Projects)
		viper.WriteConfig()
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}

func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}
