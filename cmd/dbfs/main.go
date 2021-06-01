package main

import (
	"context"
	"os"

	"github.com/andrebq/dbfs/internal/config"
	"github.com/rs/zerolog/log"
	cli "github.com/urfave/cli/v2"
)

var (
	storageConfig config.Storage
)

func combineFlags(flagSet ...[]cli.Flag) []cli.Flag {
	var ret []cli.Flag
	for _, v := range flagSet {
		ret = append(ret, v...)
	}
	return ret
}

func configureApp() *cli.App {
	var outputFlags config.Output
	app := &cli.App{
		Name:  "dbfs",
		Usage: "dbfs manages everything related to dbfs installations",
		Before: func(_ *cli.Context) error {
			outputFlags.MustConfigureGlobalLogger()
			outputFlags.MustConfigureGlobalOutput()
			return nil
		},
		Flags: combineFlags(storageConfig.AllFlags(), outputFlags.AllFlags()),
	}
	app.Commands = append(app.Commands, blobCmd())
	return app
}

func main() {
	app := configureApp()
	err := app.RunContext(context.Background(), os.Args)
	if err != nil {
		log.Error().Err(err).Msg("Failure")
		os.Exit(1)
	}
}
