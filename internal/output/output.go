package output

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/rs/zerolog/log"

	"gopkg.in/yaml.v3"
)

var (
	once        sync.Once
	formatFn    = (func(io.Writer, ...interface{}) error)(nil)
	formatOneFn = (func(io.Writer, interface{}) error)(nil)
)

func Format(output io.Writer, values ...interface{}) error {
	return formatFn(output, values)
}

func FormatOne(output io.Writer, value interface{}) error {
	return formatOneFn(output, value)
}

func MustSetFormat(name string) {
	once.Do(func() {
		switch name {
		case "json":
			formatFn = jsonOutput
			formatOneFn = jsonOneOutput
		case "json-lines":
			formatFn = jsonLinesOutput
			formatOneFn = jsonOneOutput
		case "human", "yaml":
			formatFn = humanOutput
			formatOneFn = humanOneOutput
		default:
			log.Panic().Str("output-format", name).Msg("Output format is not supported. Please check")
		}
	})
}

func humanOutput(out io.Writer, values ...interface{}) error {
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)
	return enc.Encode(values)
}

func humanOneOutput(out io.Writer, value interface{}) error {
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)
	return enc.Encode(value)
}

func jsonOutput(out io.Writer, values ...interface{}) error {
	enc := json.NewEncoder(out)
	return enc.Encode(values)
}

func jsonOneOutput(out io.Writer, value interface{}) error {
	enc := json.NewEncoder(out)
	return enc.Encode(value)
}

func jsonLinesOutput(out io.Writer, values ...interface{}) error {
	for _, v := range values {
		enc := json.NewEncoder(out)
		err := enc.Encode(v)
		if err != nil {
			return err
		}
		_, err = io.WriteString(out, "\n")
		if err != nil {
			return err
		}
	}
	return nil
}
