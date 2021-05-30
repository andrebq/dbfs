package blob

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"sync"

	"github.com/andrebq/dbfs/cas"
	"github.com/chmduquesne/rollinghash/buzhash64"
)

type (
	B struct {
		hashes sync.Pool
	}
)

const (
	// if the last 20 bytes are zero
	// take the chunk
	cutPoint = 00000000_00000000_00000000_00000000_00000000_00001111_11111111_11111111
)

var (
	chunkPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 10_000_000)
		},
	}

	emptyBuffer = bytes.NewBuffer(nil)

	scratchBufPool = sync.Pool{
		New: func() interface{} {
			return bufio.NewReaderSize(emptyBuffer, 1_000_000)
		},
	}
)

// Default uses a chunker seeded with the default value
func RandomPolinomial() (*B, error) {
	return newBlob(func() *buzhash64.Buzhash64 {
		return buzhash64.New()
	}), nil
}

// WithSeed creates a new chunker with the given seed value
func WithSeed(seed int64) (*B, error) {
	return newBlob(func() *buzhash64.Buzhash64 {
		return buzhash64.NewFromUint64Array(buzhash64.GenerateHashes(seed))
	}), nil
}

// UploadChunks reads data from r and uploads them to
// to the provided Cas object and returns the list of
// references created
func (b *B) UploadChunks(ctx context.Context, casObj *cas.C, input io.Reader) ([]cas.Ref, error) {
	// TODO: return the chunks information rather than just the refs
	// that way consumers can link chunks to the absolution position
	// in the stream, as this is useful for random access
	var window [16]byte
	var refs []cas.Ref
	n, err := input.Read(window[:])
	if err != nil {
		// input might be so short that it is less than the initial window
		// in which case, we just upload whatever bytes we just read
		if errors.Is(err, io.EOF) {
			actual := window[:n]
			ref, err := pushRef(ctx, casObj, actual)
			if err != nil {
				refs = append(refs, ref)
				return refs, nil
			}
		}
		return nil, err
	}
	block := acquireChunkBuffer()
	defer chunkPool.Put(block)
	chunk := block[:0]
	chunk = append(chunk, window[:]...)
	hasher := b.hashes.Get().(*buzhash64.Buzhash64)
	hasher.Reset()
	hasher.Write(window[:])
	defer b.hashes.Put(hasher)

	scratch := scratchBufPool.Get().(*bufio.Reader)
	scratch.Reset(input)
	defer scratchBufPool.Put(scratch)

	for b, err := scratch.ReadByte(); err == nil; b, err = scratch.ReadByte() {
		hasher.Roll(b)
		chunk = append(chunk, b)
		if (hasher.Sum64()&cutPoint) == 0 ||
			(len(chunk) == cap(block)) {
			// we either reached a cutPoint
			// or the current chunk reached the max size of a block
			ref, err := pushRef(ctx, casObj, chunk)
			if err != nil {
				return refs, err
			}
			refs = append(refs, ref)
			// reset for next chunk
			chunk = block[:0]
		}
	}

	if len(chunk) > 0 {
		// upload whatever was left
		ref, err := pushRef(ctx, casObj, chunk)
		if err != nil {
			return refs, err
		}
		refs = append(refs, ref)
		// reset for next chunk
		chunk = block[:0]
	}
	return refs, nil
}

func pushRef(ctx context.Context, casObj *cas.C, actual []byte) (cas.Ref, error) {
	return casObj.PutContent(ctx, bytes.NewBuffer(actual))
}

func newBlob(constructor func() *buzhash64.Buzhash64) *B {
	return &B{hashes: sync.Pool{
		New: func() interface{} { return constructor() },
	}}
}

func acquireChunkBuffer() []byte {
	b := chunkPool.Get().([]byte)
	b = b[0:cap(b)]
	for _, i := range b {
		b[i] = 0
	}
	return b
}
