package blob

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"testing"

	"github.com/andrebq/dbfs/cas"
	"github.com/andrebq/dbfs/internal/testutil"
)

func TestBlob(t *testing.T) {
	ctx := context.Background()
	largeRandomBuf := getRandom(t, 10, 50_000_000)
	blob, err := WithSeed(int64(0x24717b279f5337))
	if err != nil {
		t.Fatal(err)
	}
	obj, err := cas.Open(ctx, func(ctx context.Context) (cas.KV, error) {
		return testutil.MemoryBucket(ctx, t), nil
	})
	refs, err := blob.UploadChunks(ctx, obj, bytes.NewBuffer(largeRandomBuf))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("refs: %v", refs)
	t.Fail()
}

func getRandom(t *testing.T, seed int64, size int) []byte {
	// math/rand should given enough randomness to test the chunk split
	// without having to maintain large sequences of bytes in-memory
	// using a symmetric-crypto function seems overkill for this use-case
	rnd := rand.New(rand.NewSource(seed))
	buf := bytes.Buffer{}
	_, err := io.CopyN(&buf, rnd, int64(size))
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
