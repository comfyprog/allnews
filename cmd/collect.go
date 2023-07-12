package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/comfyprog/allnews/config"
	"github.com/comfyprog/allnews/feed"
	"github.com/comfyprog/allnews/storage"
	"github.com/spf13/cobra"
)

func makeNamesMap(names []string) map[string]struct{} {
	namesMap := make(map[string]struct{})
	for _, name := range names {
		namesMap[name] = struct{}{}
	}
	return namesMap
}

type dryRunner struct{}

func (*dryRunner) SaveArticles(ctx context.Context, articles []feed.Article) error {
	for _, a := range articles {
		fmt.Printf("%s\n", a.String())
	}
	return nil
}

func collect(appConfig config.Config, names []string, dryRun bool, continuous bool) {
	filterNames := len(names) > 0
	namesMap := makeNamesMap(names)

	feedGroups := make(map[string][]config.SourceConfig)
	for _, source := range appConfig.Sources {
		if filterNames {
			if _, ok := namesMap[source.Name]; !ok {
				continue

			}
		}
		if _, ok := feedGroups[source.Name]; ok {
			feedGroups[source.Name] = append(feedGroups[source.Name], source)
		} else {
			feedGroups[source.Name] = []config.SourceConfig{source}
		}
	}

	var articleStorage feed.ArticleSaver

	if dryRun {
		articleStorage = &dryRunner{}
	} else {
		var err error
		articleStorage, err = storage.NewPostgresStorage(appConfig.DbConnString)
		if err != nil {
			log.Fatal(err)
		}
	}
	feed.ProcessFeeds(context.Background(), feedGroups, articleStorage, continuous)
}

func makeCollectCmd(appConfig config.Config) *cobra.Command {
	var (
		continuous bool
		dryRun     bool
		names      []string
	)

	cmd := &cobra.Command{
		Use:   "collect",
		Short: "collects feeds",
		Long:  "Collects feeds defined in config file and stores them in the database",
		Run: func(cmd *cobra.Command, args []string) {
			collect(appConfig, names, dryRun, continuous)
		},
	}

	cmd.PersistentFlags().BoolVarP(&continuous, "continuous", "c", false, "collect feed indefinitely with interval specified in the config file")
	cmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print feed to stdout instead of saving it to the database")
	cmd.PersistentFlags().StringArrayVar(&names, "name", []string{}, "name of the feed to process (can be multiple)")
	return cmd
}
