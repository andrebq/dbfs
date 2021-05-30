// +build minio_driver

package cas

import (
	"context"
	"testing"

	"github.com/andrebq/dbfs/drivers/minio/kv"
)

func TestMinioDriver(t *testing.T) {
	// check the Makefile to get a glimpse of all environment variables
	// or read the individual options on kv
	sanityCheckCAS(t, context.Background(), func(ctx context.Context) (KV, error) {
		return kv.Connect(ctx, kv.FromEnv()...)
	})
}
