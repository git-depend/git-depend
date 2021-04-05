package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge project branches.",
	Long:  `Not currently working. Merge branches accross multiple repositories.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.Unmarshal(&cfg)
		fmt.Println("Currently not functioning.")
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}
