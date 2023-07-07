package cmd

import (
	"fmt"
	"runtime"

	"github.com/comfyprog/allnews/config"
	"github.com/spf13/cobra"
)

func makeRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "allnews",
		Short: "RSS feed aggregator",
		Long:  "Allnews is an application that gathers user-defined RSS feeds and displays them as a single timeline",
	}
}

func makeVersionCmd(config config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints version",
		Long:  "Prints program version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("version: %s, %s\n", config.Version, runtime.Version())
		},
	}
}

func Execute(config config.Config) error {
	migrateCmd := makeMigragteDbCmd(config)
	collectCmd := makeCollectCmd(config)
	serveCmd := makeServeCmd(config)
	versionCmd := makeVersionCmd(config)
	rootCmd := makeRootCmd()
	rootCmd.AddCommand(migrateCmd, collectCmd, serveCmd, versionCmd)
	return rootCmd.Execute()
}
