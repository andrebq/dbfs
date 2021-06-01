package config

import "github.com/urfave/cli/v2"

type (
	Blob struct {
		Seed int64
	}
)

func (b *Blob) AllFlags() []cli.Flag {
	return []cli.Flag{b.SeedFlag()}
}

func (b *Blob) SeedFlag() cli.Flag {
	return &cli.Int64Flag{
		Name:        "seed",
		Usage:       "Seed value used to compute the chunk codes",
		EnvVars:     []string{"DBFS_BLOB_SEED"},
		Destination: &b.Seed,
		Value:       0x24717b279f5337,
	}
}
