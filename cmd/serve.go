package cmd

import (
	"log"

	"github.com/comfyprog/allnews/config"
	"github.com/comfyprog/allnews/server"
	"github.com/comfyprog/allnews/storage"
	"github.com/spf13/cobra"
)

func makeServeCmd(config config.Config) *cobra.Command {
	var withCollect bool
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start http server",
		Long:  "Starts http server on the address and port defined in the config file",
		Run: func(cmd *cobra.Command, args []string) {
			storage, err := storage.NewPostgresStorage(config.DbConnString)
			if err != nil {
				log.Fatal(err)
			}
			log.Fatal(server.Serve(storage, config))
		},
	}

	cmd.PersistentFlags().BoolVarP(&withCollect, "with-collect", "c", false, "perform continuous feed collection in background")
	return cmd
}
