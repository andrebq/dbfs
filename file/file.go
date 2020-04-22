package file

import (
	"bytes"
	"io"
	"sort"

	"github.com/andrebq/dbfs/blob"
	"github.com/fxamacker/cbor"
)

type (
	// F represents a file which contains
	// metadata, a list of chunks, and a list
	// of child nodes, where each node
	// contains a reference to another file
	F struct {
		Meta     blob.Ref     `cbor:"1,keyasint,omitempty"`
		Chunks   RefList      `cbor:"2,keyasint,omitempty"`
		Children NamedRefList `cbor:"3,keyasint,omitempty"`
	}

	// Meta contains metadata about a given file
	Meta struct {
		// Leaf indicates that a given file should not have any children
		Leaf bool  `cbor:"1,keyasint,omitempty"`
		Size int64 `cbor:"2,keyasint,omitempty"`
	}

	// RefList is a ordered list of references
	RefList []blob.Ref

	// NamedRef attaches a name to a reference
	NamedRef struct {
		_    struct{} `cbor:",toarray"`
		Name string
		Ref  blob.Ref
	}

	// NamedRefList is a sorted list of named refs (sorted by name)
	NamedRefList []NamedRef
)

const (
	B              = 1
	KB             = 1000 * B
	MB             = 1000 * KB
	IdealChunkSize = 1 * MB
)

func (f *F) ToBlob() []byte {
	b := &bytes.Buffer{}
	f.WriteTo(b)
	return b.Bytes()
}

func (f *F) WriteTo(w io.Writer) error {
	enc := cbor.NewEncoder(w, cbor.CTAP2EncOptions())
	return enc.Encode(f)
}

func (f *F) ReadFrom(in io.Reader) error {
	dec := cbor.NewDecoder(in)
	return dec.Decode(f)
}

func (nr NamedRefList) MarshalCBOR() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := cbor.NewEncoder(buf, cbor.CTAP2EncOptions())
	sort.Sort(nr)
	err := enc.Encode([]NamedRef(nr))
	return buf.Bytes(), err
}

func (nr NamedRefList) Less(i, j int) bool {
	return nr[i].Name < nr[j].Name
}
func (nr NamedRefList) Len() int { return len(nr) }
func (nr NamedRefList) Swap(i, j int) {
	nr[i], nr[j] = nr[j], nr[i]
}

func (m *Meta) ToBlob() []byte {
	buf, _ := cbor.Marshal(m, cbor.CTAP2EncOptions())
	return buf
}
