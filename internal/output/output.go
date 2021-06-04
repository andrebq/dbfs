package output

import (
	"encoding/json"
	"io"
	"reflect"
	"sync"

	"github.com/andrebq/dbfs/internal/tuple"
	"github.com/rs/zerolog/log"

	"gopkg.in/yaml.v3"
)

var (
	once        sync.Once
	formatFn    = (func(io.Writer, interface{}) error)(nil)
	formatOneFn = (func(io.Writer, interface{}) error)(nil)
)

func Format(output io.Writer, values interface{}) error {
	return formatFn(output, values)
}

func FormatItems(output io.Writer, items ...interface{}) error {
	for _, v := range items {
		err := Format(output, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// func FormatOne(output io.Writer, value interface{}) error {
// 	return formatOneFn(output, value)
// }

// Spread value as a list of elements, if val is a Slice object
// than it will copy all items to from val to a new auxiliary list
//
// If val is not a Slice object, then it return a slice object
// with just one element.
//
// This process wastes a lot of memory and generates a bunch
// of points, but most of the time this happens right before
// the end of the process, so GC impact should be minimal
func Spread(val interface{}) []interface{} {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Slice {
		tmp := make([]interface{}, v.Len())
		for i := range tmp {
			tmp[i] = v.Index(i).Interface()
		}
		return tmp
	}
	return []interface{}{val}
}

func MustSetFormat(name string) {
	once.Do(func() {
		switch name {
		case "json":
			formatFn = jsonOutput
			formatOneFn = jsonOneOutput
		// case "json-lines":
		// 	formatFn = jsonLinesOutput
		// 	formatOneFn = jsonOneOutput
		case "human", "yaml":
			formatFn = humanOutput
			formatOneFn = humanOneOutput
		case "binary":
			formatFn = binaryOutput
			formatOneFn = binaryOneOutput
		default:
			log.Panic().Str("output-format", name).Msg("Output format is not supported. Please check")
		}
	})
}

func humanOutput(out io.Writer, values interface{}) error {
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)
	return enc.Encode(values)
}

func humanOneOutput(out io.Writer, value interface{}) error {
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)
	return enc.Encode(value)
}

func jsonOutput(out io.Writer, values interface{}) error {
	enc := json.NewEncoder(out)
	return enc.Encode(values)
}

func jsonOneOutput(out io.Writer, value interface{}) error {
	enc := json.NewEncoder(out)
	return enc.Encode(value)
}

// func jsonLinesOutput(out io.Writer, values interface{}) error {
// 	for _, v := range values {
// 		enc := json.NewEncoder(out)
// 		err := enc.Encode(v)
// 		if err != nil {
// 			return err
// 		}
// 		_, err = io.WriteString(out, "\n")
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func binaryOutput(out io.Writer, values interface{}) error {
	return binaryOneOutput(out, values)
}

func binaryOneOutput(out io.Writer, value interface{}) error {
	switch value := value.(type) {
	case interface{ MarshalBinary() ([]byte, error) }:
		buf, err := value.MarshalBinary()
		if err != nil {
			return err
		}
		_, err = out.Write(buf)
		return err
	case tuple.Content:
		buf, err := tuple.MarshalBinary(value)
		if err != nil {
			return err
		}
		_, err = out.Write(buf)
		return err
	default:
		buf, err := tuple.MarshalBinary(tuple.Indexed{value})
		if err != nil {
			return err
		}
		_, err = out.Write(buf)
		return err
	}
}
