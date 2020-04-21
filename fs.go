package dbfs

import (
	"io"

	"github.com/andrebq/dbfs/blob"
	"github.com/andrebq/dbfs/file"
	"github.com/andrebq/dbfs/seek"
)

// WriteFile copies r to f and storing all intermediate references into cas.
// the content of r is split into chunks of 1MB but if the file is too large,
// the file object might end up a value larger than 1MB
func WriteFile(cas blob.CAS, r io.Reader) (blob.Ref, error) {
	f := file.F{}
	buf := make([]byte, file.IdealChunkSize)
	eof := false
	for !eof {
		n, err := r.Read(buf)
		if err == io.EOF {
			eof = true
		} else if err != nil {
			return blob.Ref{}, err
		}
		if n == 0 {
			break
		}
		ref, err := cas.Put(seek.NewReadOnlyBuffer(buf[:n]))
		if err != nil {
			return blob.Ref{}, err
		}
		f.Chunks = append(f.Chunks, ref)
	}
	return cas.Put(seek.NewReadOnlyBuffer(f.ToBlob()))
}
