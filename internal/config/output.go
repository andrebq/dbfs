package config

import (
	"github.com/andrebq/dbfs/internal/output"
	"github.com/urfave/cli/v2"
)

type (
	Output struct {
		Format    string
		LogFormat string
		Level     string
	}
)

func (o *Output) AllFlags() []cli.Flag {
	return []cli.Flag{
		o.OutputFormatFlag("o", "format"),
		o.LogFormatFlag(),
		o.LogLevelFlag(),
	}
}

func (o *Output) OutputFormatFlag(aliases ...string) cli.Flag {
	return &cli.StringFlag{
		Name:        "stdout-output-format",
		Usage:       "Output format used by the tool when writing structures to stdout",
		Value:       "human",
		Aliases:     aliases,
		EnvVars:     []string{"DBFS_OUTPUT_FORMAT"},
		Destination: &o.Format,
	}
}

func (o *Output) LogFormatFlag() cli.Flag {
	return &cli.StringFlag{
		Name:        "log-format",
		Usage:       "Format used by any log message written by the application",
		Value:       "json",
		EnvVars:     []string{"DBFS_LOG_FORMAT"},
		Destination: &o.LogFormat,
	}
}

func (o *Output) LogLevelFlag() cli.Flag {
	return &cli.StringFlag{
		Name:        "log-level",
		Usage:       "Log level",
		Value:       "info",
		EnvVars:     []string{"DBFS_LOG_LEVEL"},
		Destination: &o.Level,
	}
}

func (o *Output) MustConfigureGlobalOutput() {
	output.MustSetFormat(o.Format)
}
