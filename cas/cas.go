package cas

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/google/uuid"
)

type (
	// C implements the CAS abstraction on top of a
	// S3 compatible storage
	C struct {
		dataTable KV

		dataPath, tempPath string
		rootTmpUUIDs       uuid.UUID
		objCount           uint64

		hexDirCount int
	}

	// Ref contains the binary value of the sha256 hash which identifies
	// any object
	Ref [32]byte

	// NewTable should return a KV object which is used to
	// store the items
	//
	// cas package is responsible for closing the KV
	NewTable func(context.Context) (KV, error)
)

var (
	uuidCAS       = uuid.NewSHA1(uuid.NameSpaceOID, []byte("cas"))
	uuidTmpBucket = uuid.NewSHA1(uuidCAS, []byte("temporary-buckets"))
)

// Open a new CAS store using newBucket to acquire the remote item
func Open(ctx context.Context, newBucket NewTable) (*C, error) {
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
		dataTable:    bucket,
		rootTmpUUIDs: tmpBucket,
		hexDirCount:  4,
	}, nil
}

// PutContent writes content to a temporary object and later copies that object
// to the final path under the sha256 hash.
//
// This avoids reading the object twice but might incur in costs
// on S3-like services, even though the temporary object is alive
// for a short period of time.
//
// This function does not perform deduplication, so if the content already
// exists, it will be re-uploaded. The Copy/Move operation wont be executed
// in this scenario.
//
// If the provided KV object implementes the Mover interface, then instead
// of Copy/Delete cas will use the Move operation.
func (c *C) PutContent(ctx context.Context, content io.Reader) (Ref, error) {
	var counterInBytes [8]byte
	c.objCount++
	uint64Bytes(&counterInBytes, c.objCount)
	tmpIdentity := uuid.NewSHA1(c.rootTmpUUIDs, counterInBytes[:])
	tmpPath := path.Join(c.tempPath, tmpIdentity.String())
	var ref Ref
	rc := RefCalculator(&ref, content)
	defer rc.Close()
	_, err := c.dataTable.Write(ctx, tmpPath, rc)
	if err != nil {
		return Ref{}, err
	}
	finalPath := path.Join(c.dataPath, ref.HexPath(c.hexDirCount))
	if exists, _ := c.dataTable.Exists(ctx, finalPath); exists {
		return ref, nil
	}
	err = move(ctx, c.dataTable, finalPath, tmpPath)
	if err != nil {
		return Ref{}, fmt.Errorf("unable to copy %v to %v, cause: %w", tmpPath, finalPath, err)
	}
	return ref, nil
}

// Exists returns true if the ref already exists
func (c *C) Exists(ctx context.Context, ref Ref) (bool, error) {
	return c.dataTable.Exists(ctx, path.Join(c.dataPath, ref.HexPath(c.hexDirCount)))
}

// Get writes the object at ref to the given output
// it returns the underlying KV error without any modification
func (c *C) GetContent(ctx context.Context, w io.Writer, ref Ref) error {
	finalPath := path.Join(c.dataPath, ref.HexPath(c.hexDirCount))
	_, err := c.dataTable.Read(ctx, w, finalPath)
	return err
}

// Close the underlying bucket
func (c *C) Close() error {
	errData := c.dataTable.Close()
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
