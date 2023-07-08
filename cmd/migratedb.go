package cmd

import (
	"github.com/comfyprog/allnews/config"
	"github.com/comfyprog/allnews/storage"
	"github.com/spf13/cobra"
)

func makeMigragteDbCmd(config config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "migratedb",
		Short: "Apply database migrations",
		Long:  "migratedb command tries to setup the database to be usable with this version of allnews",
		Run: func(cmd *cobra.Command, args []string) {
			storage.Migrate(config.DbConnString)
		},
	}
}
