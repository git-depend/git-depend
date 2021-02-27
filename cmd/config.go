package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Write to the configuration file.",
	Long:  `Add your name and email address to a configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.Unmarshal(&cfg)
		fmt.Println("Writing config...")
		viper.WriteConfig()
		fmt.Println(cfg.String())
	},
}

func init() {
	configCmd.Flags().StringVar(&cfg.Author, "author", "", "Author name")
	viper.BindPFlag("author", configCmd.Flags().Lookup("author"))

	configCmd.Flags().StringVar(&cfg.Email, "email", "", "Email address")
	viper.BindPFlag("email", configCmd.Flags().Lookup("email"))

	rootCmd.AddCommand(configCmd)
}
