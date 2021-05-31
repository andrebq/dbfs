package cas

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"io/ioutil"
	"strings"
	"sync"
)

type (
	// refCalculator implements the reader interface and allows
	// computes the hash of the content read from the underlying
	// reader
	//
	// The value is updated when an EOF (or any other error)
	// is returned from the underlying reader
	refCalculator struct {
		actual io.Reader
		out    *Ref
		hasher hash.Hash
	}

	// RollingRef computes a ref value as bytes are written to it
	// it is just a wraper over a hash.Hash
	RollingRef interface {
		io.WriteCloser
		io.ByteWriter
		Ref() Ref
		Reset()
	}

	rollingRef struct {
		hasher  hash.Hash
		onebyte []byte
		ref     Ref
	}
)

var (
	sha256Pool = sync.Pool{
		New: func() interface{} { return sha256.New() },
	}
)

// Returns the hex encoded path with the first hex-bytes
// used as directories
//
// n MUST be less than 32
func (r Ref) HexPath(n int) string {
	hexstr := hex.EncodeToString(r[:])
	if n == 0 {
		return hexstr
	}
	parts := make([]string, n, n+1)
	for i := range parts {
		parts[i] = hexstr[:2]
		hexstr = hexstr[2:]
	}
	// append the tail
	parts = append(parts, hexstr)
	return strings.Join(parts, "/")
}

// String returns the hex encoding of this object
func (r Ref) String() string {
	return hex.EncodeToString(r[:])
}

// RefCaclulator returns a reader that computes the hash from
// the given content as consumers read data.
//
// The Ref *pointer is updated when content.Read returns 0
// bytes or an error is found, including io.EOF
func RefCalculator(out *Ref, content io.Reader) io.ReadCloser {
	h := sha256Pool.Get().(hash.Hash)
	h.Reset()
	return &refCalculator{
		out:    out,
		actual: content,
		hasher: h,
	}
}

// NewRollingRef configures a new rolling hash object
func NewRollingRef() RollingRef {
	hasher := sha256Pool.Get().(hash.Hash)
	hasher.Reset()
	return &rollingRef{
		hasher:  hasher,
		onebyte: make([]byte, 1),
	}
}

// PrecomputeHashBytes returns the expected Ref value for the
// given set of bytes
func PrecomputeHashBytes(buf []byte) (ref Ref) {
	rc := RefCalculator(&ref, bytes.NewBuffer(buf))
	io.Copy(ioutil.Discard, rc)
	return
}

func (r *refCalculator) Read(buf []byte) (int, error) {
	n, err := r.actual.Read(buf)
	if n == 0 || err != nil {
		r.hasher.Sum((*r.out)[:0])
	} else {
		r.hasher.Write(buf[:n])
	}
	return n, err
}

func (r *refCalculator) Close() error {
	sha256Pool.Put(r.hasher)
	if closer, ok := r.actual.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (rr *rollingRef) Write(b []byte) (int, error) {
	return rr.hasher.Write(b)
}

func (rr *rollingRef) WriteByte(b byte) error {
	rr.onebyte[0] = b
	_, err := rr.hasher.Write(rr.onebyte)
	return err
}

func (rr *rollingRef) Ref() Ref {
	rr.hasher.Sum(rr.ref[:0])
	return rr.ref
}

func (rr *rollingRef) Close() error {
	sha256Pool.Put(rr.hasher)
	return nil
}

func (rr *rollingRef) Reset() {
	rr.hasher.Reset()
}
