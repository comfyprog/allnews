package main

import (
	"os"

	"github.com/comfyprog/allnews/cmd"
	"github.com/comfyprog/allnews/config"
)

const version = "0.0.1"

func main() {
	confFilename, ok := os.LookupEnv("ALLNEWS_CONFIG")
	if !ok {
		confFilename = "config.yml"
	}

	config, err := config.Get(confFilename, version)
	if err != nil {
		panic(err)
	}

	cmd.Execute(config)
}
