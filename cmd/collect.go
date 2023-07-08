package cmd

import (
	"fmt"

	"github.com/comfyprog/allnews/config"
	"github.com/spf13/cobra"
)

func makeCollectCmd(config config.Config) *cobra.Command {
	var continuous bool
	cmd := &cobra.Command{
		Use:   "collect",
		Short: "collects feeds",
		Long:  "Collects feeds defined in config file and stores them in the database",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("TODO: collect (continuously: %v), args: %v\n", continuous, args)
			fmt.Printf("%+v\n", config)
		},
	}

	cmd.PersistentFlags().BoolVarP(&continuous, "continuous", "c", false, "collect feed indefinitely with interval specified in the config file")
	return cmd
}
