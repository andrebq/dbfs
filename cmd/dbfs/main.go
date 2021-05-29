package main

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	cli "github.com/urfave/cli/v2"
)

func configureApp() *cli.App {
	return &cli.App{
		Name:  "dbfs",
		Usage: "dbfs manages everything related to dbfs installations",
	}
}

func main() {
	log.Info().Strs("args", os.Args[1:]).Str("binary", os.Args[0]).Msg("Starting dbfs")
	app := configureApp()
	err := app.RunContext(context.Background(), os.Args)
	if err != nil {
		log.Error().Err(err).Msg("Failure")
		os.Exit(1)
	}
}
