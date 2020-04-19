package main

import (
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/andrebq/dbfs"
	"github.com/andrebq/dbfs/seek"
)

var (
	action = flag.String("action", "put", `Operation to execute
	get: to return the content of a reference (if available)
	put: to read from stdin, store in the database and return the reference
	list: to list all references available (follows the directory transversal order)`)

	encoding = flag.String("encoding", "hex", `Encoding to use when reading/writing blob refs
	hex: to write hexadecimal
	base64: standard base64 encoding with padding
	base64u: url-safe base64 encoding with padding
	raw: raw bytes
	
	Apart from raw (which uses 0 as separator), all other encodings include a new line after each reference`)

	dir = flag.String("dir", "", "Directory to use for blob storage")
	ns  = flag.String("ns", "", "Namespace used to isolate blob stores from each other")
)

func main() {
	flag.Parse()
	cas, err := dbfs.OpenFileCAS(*dir, *ns)
	if err != nil {
		log.Fatalf("Unable to open CAS file with the given parameter: %v", err)
	}
	switch *action {
	case "get":
		doGet()
	case "put":
		doPut(cas)
	case "list":
		doList(cas)
	default:
		flag.Usage()
	}
}

func doList(cas dbfs.CAS) {
	out := make(chan dbfs.Ref, 1000)
	err := make(chan error, 1)
	go cas.List(out, err)
	for {
		select {
		case r, valid := <-out:
			if !valid {
				return
			}
			writeRef(os.Stdout, &r)
		case listErr, valid := <-err:
			if !valid {
				return
			}
			log.Fatalf("Error listing references: %v", listErr)
		}
	}
}

func writeRef(out io.Writer, ref *dbfs.Ref) error {
	var err error
	switch *encoding {
	case "base64u":
		_, err = fmt.Fprintf(out, "%v\n", base64.URLEncoding.EncodeToString(ref[:]))
	case "base64":
		_, err = fmt.Fprintf(out, "%v\n", base64.StdEncoding.EncodeToString(ref[:]))
	case "hex":
		_, err = fmt.Fprintf(out, "%v\n", hex.EncodeToString(ref[:]))
	case "raw":
		_, err = out.Write(ref[:])
		if err != nil {
			return err
		}
		_, err = out.Write([]byte{0})
	}
	return err
}

func doGet() {
	log.Fatal("get not implemented")
}

func doPut(cas dbfs.CAS) {
	seekable := seek.CopyToTemp(os.Stdin)
	defer seekable.Close()
	ref, err := cas.Put(seekable)
	if err != nil {
		log.Fatalf("cas.Put failed with %v", err)
	}
	writeRef(os.Stdout, &ref)
}
