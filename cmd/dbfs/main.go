package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/andrebq/dbfs"
	"github.com/andrebq/dbfs/blob"
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

	outEncoding = flag.String("outEncoding", "raw", `Encoding used to print data to stdout, same options as -encoding`)

	dir = flag.String("dir", "", "Directory to use for blob storage")
	ns  = flag.String("ns", "", "Namespace used to isolate blob stores from each other")
)

func main() {
	flag.Parse()
	cas, err := blob.OpenFileCAS(*dir, *ns)
	if err != nil {
		log.Fatalf("Unable to open CAS file with the given parameter: %v", err)
	}
	switch *action {
	case "get":
		doGet(cas)
	case "put":
		doPut(cas)
	case "list":
		doList(cas)
	default:
		flag.Usage()
	}
}

func doList(cas blob.CAS) {
	out := make(chan blob.Ref, 1000)
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

func writeRef(out io.Writer, ref *blob.Ref) error {
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

func readRef(ref *blob.Ref, in io.Reader) error {
	sc := bufio.NewScanner(in)
	for sc.Scan() {
		switch *encoding {
		case "hex":
			_, err := hex.Decode(ref[:], sc.Bytes())
			return err
		default:
			return errors.New("encoding not yet implemented for read operations")
		}
	}
	return sc.Err()
}

func doGet(cas blob.CAS) {
	var ref blob.Ref
	err := readRef(&ref, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	out := &bytes.Buffer{}
	err = cas.Copy(out, ref)
	if err != nil {
		log.Fatal(err)
	}
	switch *outEncoding {
	case "hex":
		os.Stdout.WriteString(hex.EncodeToString(out.Bytes()))
	case "raw":
		os.Stdout.Write(out.Bytes())
	}
}

func doPut(cas blob.CAS) {
	ref, err := dbfs.WriteFile(cas, os.Stdin)
	if err != nil {
		log.Fatal("Unable to save file to dbfs", err)
	}
	writeRef(os.Stdout, &ref)
}
