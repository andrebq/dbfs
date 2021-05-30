package kv

import (
	"context"
	"io"

	"gocloud.dev/blob"
)

type (
	Bucket struct {
		actual *blob.Bucket
	}
)

// Connect to a Go Cloud blob bucket
func Connect(ctx context.Context, url string) (*Bucket, error) {
	bucket, err := blob.OpenBucket(ctx, url)
	if err != nil {
		return nil, err
	}
	return &Bucket{actual: bucket}, nil
}

func (b *Bucket) Close() error { return b.actual.Close() }
func (b *Bucket) Copy(ctx context.Context, to, from string) error {
	return b.actual.Copy(ctx, to, from, nil)
}
func (b *Bucket) Delete(ctx context.Context, key string) error {
	return b.actual.Delete(ctx, key)
}
func (b *Bucket) Write(ctx context.Context, path string, input io.Reader) (int64, error) {
	writer, err := b.actual.NewWriter(ctx, path, nil)
	if err != nil {
		return 0, err
	}
	defer writer.Close()
	return io.Copy(writer, input)
}
func (b *Bucket) Read(ctx context.Context, w io.Writer, path string) (int64, error) {
	reader, err := b.actual.NewReader(ctx, path, nil)
	if err != nil {
		return 0, err
	}
	defer reader.Close()
	return io.Copy(w, reader)
}
func (b *Bucket) Exists(ctx context.Context, path string) (bool, error) {
	return b.actual.Exists(ctx, path)
}
