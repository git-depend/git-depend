package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/git-depend/git-depend/pkg/depend"
	"github.com/git-depend/git-depend/pkg/git"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commits a dependency structure.",
	Long:  `Creates a tree of dependencies and attempts to commit them.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.Unmarshal(&cfg)

		if cfg.Author == "" {
			fmt.Println("Please set the author in the config.")
			fmt.Println("Hint: config --help")
			os.Exit(1)
		}

		if cfg.Email == "" {
			fmt.Println("Please set the email in the config.")
			fmt.Println("Hint: config --help")
			os.Exit(1)
		}

		// Create the cache
		cache_dir := getCacheDir()
		cache, err := git.NewCache(cache_dir)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Create the requests
		reqs := make([]*depend.Request, len(cfg.Projects))
		urls := make([]string, len(cfg.Projects))
		for i := 0; i < len(cfg.Projects); i++ {
			repo, branch := getRepoBranch(cfg.Projects[i])
			reqs[i] = depend.NewRequest(
				repo,
				branch,
				cfg.Author,
				cfg.Email,
				time.Now().String(),
				nil)
			urls[i] = repo
		}

		if _, err := cache.CloneOrUpdateMany(urls); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		ref := "git-depend"
		for _, req := range reqs {
			if err := req.Write(cache, ref); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		for _, url := range urls {
			request := &depend.Request{}
			request.ReadFromUrl(cache, url, ref)
			// Prints pretty structs
			pretty, _ := json.MarshalIndent(request, "", "\t")
			fmt.Println(string(pretty))
		}
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}

// Takes a "url:branch" and returns the (url,branch)
func getRepoBranch(s string) (string, string) {
	split := strings.Split(s, ":")
	var repo []string
	for i := 0; i < len(split)-1; i++ {
		repo = append(repo, split[i])
	}
	return strings.Join(repo, ":"), split[len(split)-1]
}
