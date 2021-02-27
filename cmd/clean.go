package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean the cache.",
	Long:  `Deletes everything in the cache.`,
	Run: func(cmd *cobra.Command, args []string) {
		cache_dir := getCacheDir()
		fmt.Println("Cleaning: ", cache_dir)
		if err := os.RemoveAll(getCacheDir()); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
