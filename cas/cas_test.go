package cas

import (
	"bytes"
	"context"
	"encoding/hex"
	"testing"

	"github.com/andrebq/dbfs/internal/testutil"
	"gocloud.dev/blob"
)

func TestCAS(t *testing.T) {
	ctx := context.Background()

	cas, err := Open(ctx, func(ctx context.Context) (*blob.Bucket, error) {
		return testutil.MemoryBucket(ctx, t), nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// echo -n abc123 | shasum -a 256
	expectedRef, _ := hex.DecodeString("6ca13d52ca70c883e0f0bb101e425a89e8624de51db2d2392593af6a84118090")
	ref, err := cas.PutContent(ctx, bytes.NewBufferString("abc123"))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(ref[:], expectedRef[:]) {
		t.Errorf("Expecting hash to be: %v got %v", hex.EncodeToString(expectedRef), hex.EncodeToString(ref[:]))
	}

	buf := &bytes.Buffer{}
	if err := cas.GetContent(ctx, buf, ref); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(buf.Bytes(), []byte("abc123")) {
		t.Errorf("Unexpected content from buffer: %v", buf.String())
	}
}
