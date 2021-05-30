package testutil

import (
	"context"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
)

// TestBucket returns a bucket that holds its content
// in the process memory space
func MemoryBucket(ctx context.Context, t interface{ Fatal(...interface{}) }) *blob.Bucket {
	bucket, err := blob.OpenBucket(ctx, "mem://")
	if err != nil {
		t.Fatal(err)
	}
	return bucket
}
