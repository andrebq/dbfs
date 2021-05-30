package testutil

import (
	"context"

	"github.com/andrebq/dbfs/drivers/gcloud/kv"
	_ "gocloud.dev/blob/memblob"
)

// TestBucket returns a bucket that holds its content
// in the process memory space
func MemoryBucket(ctx context.Context, t interface{ Fatal(...interface{}) }) *kv.Bucket {
	bucketKV, err := kv.Connect(ctx, "mem://")
	if err != nil {
		t.Fatal(err)
	}
	return bucketKV
}
