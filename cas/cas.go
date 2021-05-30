package cas

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/google/uuid"
	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"
)

type (
	// C implements the CAS abstraction on top of a
	// S3 compatible storage
	C struct {
		dataBucket *blob.Bucket

		dataPath, tempPath string
		rootTmpUUIDs       uuid.UUID
		objCount           uint64

		hexDirCount int
	}

	// Ref contains the binary value of the sha256 hash which identifies
	// any object
	Ref [32]byte

	// NewBucket should return a blob.Bucket object with the provided
	// prefix.
	//
	// cas package is responsible for closing the bucket
	NewBucket func(context.Context) (*blob.Bucket, error)
)

var (
	uuidCAS       = uuid.NewSHA1(uuid.NameSpaceOID, []byte("cas"))
	uuidTmpBucket = uuid.NewSHA1(uuidCAS, []byte("temporary-buckets"))
)

// Open a new CAS store using newBucket to acquire the remote item
func Open(ctx context.Context, newBucket NewBucket) (*C, error) {
	var c C
	c.dataPath = path.Join("data")
	c.tempPath = path.Join("tmp")

	bucket, err := newBucket(ctx)
	if err != nil {
		return nil, err
	}
	var nowInBytes [8]byte
	int64Bytes(&nowInBytes, time.Now().Unix())
	tmpBucket := uuid.NewSHA1(uuidTmpBucket, nowInBytes[:])

	return &C{
		dataBucket:   bucket,
		rootTmpUUIDs: tmpBucket,
		hexDirCount:  4,
	}, nil
}

// PutContent writes content to a temporary object and later copies that object
// to the final path under the sha256 hash.
//
// This avoids reading the object twice but might incur in costs on S3 service,
// the temporary object remains alive only for a short period of time.
//
// TODO: use interfaces and check if the Move operation is supproted, in most
// providers, moving is cheaper tha copy/delete.
func (c *C) PutContent(ctx context.Context, content io.Reader) (Ref, error) {
	var counterInBytes [8]byte
	c.objCount++
	uint64Bytes(&counterInBytes, c.objCount)
	tmpIdentity := uuid.NewSHA1(c.rootTmpUUIDs, counterInBytes[:])
	tmpPath := path.Join(c.tempPath, tmpIdentity.String())
	writer, err := c.dataBucket.NewWriter(ctx, tmpPath, &blob.WriterOptions{})
	if err != nil {
		return Ref{}, err
	}
	var ref Ref
	_, err = writer.ReadFrom(RefCalculator(&ref, content))
	if err != nil {
		writer.Close()
		return Ref{}, err
	}
	writer.Close()

	finalPath := path.Join(c.dataPath, ref.HexPath(c.hexDirCount))

	err = c.dataBucket.Copy(ctx, finalPath, tmpPath, &blob.CopyOptions{})
	if err != nil {
		return Ref{}, fmt.Errorf("unable to copy %v to %v, cause: %w", tmpPath, finalPath, err)
	}
	return ref, nil
}

// Get writes the object at ref to the given output
// or returns ErrNotFound if the reference is invalid
func (c *C) GetContent(ctx context.Context, w io.Writer, ref Ref) error {
	finalPath := path.Join(c.dataPath, ref.HexPath(c.hexDirCount))
	reader, err := c.dataBucket.NewReader(ctx, finalPath, &blob.ReaderOptions{})
	if err != nil {
		switch gcerrors.Code(err) {
		case gcerrors.NotFound:
			return ErrNotFound
		default:
			return err
		}
	}
	defer reader.Close()
	_, err = io.Copy(w, reader)
	if err != nil {
		return err
	}
	return nil
}

// Close the underlying bucket
func (c *C) Close() error {
	errData := c.dataBucket.Close()
	if errData != nil {
		return fmt.Errorf("unable to close data bucket, cause: %w", errData)
	}
	return nil
}

func int64Bytes(out *[8]byte, in int64) {
	binary.BigEndian.PutUint64((*out)[:], uint64(in))
}

func uint64Bytes(out *[8]byte, in uint64) {
	binary.BigEndian.PutUint64((*out)[:], uint64(in))
}
