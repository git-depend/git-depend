package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type config struct {
	Author   string
	Email    string
	Projects []string
}

func (cfg *config) String() string {
	s, _ := json.MarshalIndent(cfg, "", "\t")
	return string(s)
}

var cfg config
var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-dep",
	Short: "Manage multiple git projects.",
	Long: `Manage multiple git projects.
			For more information, go to:
          	github.com/git-depend/git-depend`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.git-depend)")
}

// initConfig reads in config file.
func initConfig() {
	viper.SetConfigType("toml")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		configName := ".git-depend"
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Search config in home directory with name ".git-depend.toml".
		viper.AddConfigPath(home)
		viper.SetConfigName(configName)
		// Create a configfile if one doesn't exist.
		viper.SafeWriteConfig()
	}

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Using config file:", viper.ConfigFileUsed())
}

func getCacheDir() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return path.Join(home, ".cache", "git-depend")
}
