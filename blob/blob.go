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

	// Chunk contains a sequence of bytes whose hash value
	// matches the CutPoint
	Chunk struct {
		Start int64
		End   int64
		Size  int
		Sum   uint64
		// Short indicates if the chunk was so short that
		// rolling hash could not be computed for it
		//
		// Any block with less than 16 bytes is considered too short
		Short bool

		// Ref contains the reference that would be used to identify
		// this specific chunk of data
		Ref cas.Ref
	}
)

const (
	// CutPoint contains the number of bits from a hash that should be 0
	// to consider that block as one individual chunk
	//
	// Basically
	//
	// RollingHashOfChunk & CutPoint == 0
	CutPoint = 00000000_00000000_00000000_00000000_00000000_00001111_11111111_11111111

	maxChunkSize = 10_000_000
)

var (
	chunkPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, maxChunkSize)
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

// Chunks takes the given input and splits it into chunks
// using rolling hash, chunks are written to the channel,
// although the actual data is not sent (thus avoiding allocating)
// lots of objects for large streams
func (b *B) Chunks(ctx context.Context, input io.Reader) ([]Chunk, error) {
	// TODO: return the chunks information rather than just the refs
	// that way consumers can link chunks to the absolution position
	// in the stream, as this is useful for random access
	var window [16]byte
	var chunks []Chunk
	var lastChunk Chunk
	n, err := input.Read(window[:])
	if err != nil {
		// input might be so short that it is less than the initial window
		// in which case, we just upload whatever bytes we just read
		if errors.Is(err, io.EOF) {
			rr := cas.NewRollingRef()
			rr.Write(window[:n])
			lastChunk.Ref = rr.Ref()
			rr.Close()
			lastChunk.Size = n
			lastChunk.End = lastChunk.Start + int64(lastChunk.Size)
			lastChunk.Short = n < cap(window)
			if !lastChunk.Short {
				hasher := b.hashes.Get().(*buzhash64.Buzhash64)
				hasher.Reset()
				hasher.Write(window[:])
				lastChunk.Sum = hasher.Sum64()
				b.hashes.Put(hasher)
			}
			chunks = append(chunks, lastChunk)
			return chunks, nil
		}
		return nil, err
	}
	rr := cas.NewRollingRef()
	defer rr.Close()
	rr.Write(window[:])
	hasher := b.hashes.Get().(*buzhash64.Buzhash64)
	hasher.Reset()
	hasher.Write(window[:])
	defer b.hashes.Put(hasher)
	lastChunk.Size = len(window)

	scratch := scratchBufPool.Get().(*bufio.Reader)
	scratch.Reset(input)
	defer scratchBufPool.Put(scratch)

	for b, err := scratch.ReadByte(); err == nil; b, err = scratch.ReadByte() {
		hasher.Roll(b)
		rr.WriteByte(b)
		lastChunk.Size++
		if (hasher.Sum64()&CutPoint) == 0 ||
			lastChunk.Size == maxChunkSize {
			lastChunk.End = lastChunk.Start + int64(lastChunk.Size)
			lastChunk.Ref = rr.Ref()
			chunks = append(chunks, lastChunk)
			rr.Reset()
			lastChunk.Size = 0
			lastChunk.Start = lastChunk.End
			lastChunk.End = lastChunk.Start
		}
	}

	if errors.Is(err, io.EOF) {
		err = nil
	}

	if err != nil {
		return nil, err
	}

	if lastChunk.Size > 0 {
		lastChunk.End = lastChunk.Start + int64(lastChunk.Size)
		lastChunk.Sum = hasher.Sum64()
		lastChunk.Ref = rr.Ref()
		chunks = append(chunks, lastChunk)
	}
	return chunks, nil
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
		if (hasher.Sum64()&CutPoint) == 0 ||
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
