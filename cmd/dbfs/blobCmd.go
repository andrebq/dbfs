package main

import (
	"errors"
	"io"
	"os"

	"github.com/andrebq/dbfs/blob"
	"github.com/andrebq/dbfs/internal/config"
	"github.com/andrebq/dbfs/internal/output"
	cli "github.com/urfave/cli/v2"
)

func blobCmd() *cli.Command {
	var cfg config.Blob
	return &cli.Command{
		Name:  "blob",
		Usage: "Sub-command to interact with the blob chunk sub-system",
		Flags: cfg.AllFlags(),
		Subcommands: []*cli.Command{
			blobChunkSubcommand(&cfg),
		},
	}
}

func blobChunkSubcommand(cfg *config.Blob) *cli.Command {
	var fileName string
	return &cli.Command{
		Name:  "chunks",
		Usage: "Take a file input (or stdin by default) and writes to stdout the chunks for that file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "file",
				Aliases:     []string{"i"},
				Value:       "-",
				Usage:       "File to read as input (stdin is default)",
				Destination: &fileName,
			},
		},
		Action: func(appCtx *cli.Context) error {
			b, err := blob.WithSeed(cfg.Seed)
			if err != nil {
				return err
			}
			var file io.Reader
			switch fileName {
			case "":
				return errors.New("fileName flag cannot be empty")
			case "-":
				file = os.Stdin
			default:
				fd, err := os.Open(fileName)
				if err != nil {
					return err
				}
				defer fd.Close()
				file = fd
			}
			chunks, err := b.Chunks(appCtx.Context, file)
			if err != nil {
				return err
			}
			return output.FormatOne(os.Stdout, struct {
				Chunks []blob.Chunk `json:"chunks" yaml:"chunks"`
			}{
				Chunks: chunks,
			})
		},
	}
}
