package kv

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type (
	Bucket struct {
		cli    *minio.Client
		bucket string
		region string
	}

	countReader struct {
		total  int64
		actual io.Reader
	}

	config struct {
		endpoint        string
		accessKeyID     string
		secretAccessKey string
		secretToken     string
		bucket          string
		region          string
	}

	Option func(cfg *config) error
)

// FromEnv returns the list of options loaded from the environment
// if an option from the environment is already present then its value
// will be ignored.
func FromEnv() []Option {
	return []Option{
		EndpointFromEnv,
		AccessKeyIDFromEnv,
		SecretAccessKeyFromEnv,
		TokenFromEnv,
		BucketFromEnv,
		RegionFromEnv,
	}
}

func RegionPtr(value *string) Option {
	return func(cfg *config) error {
		cfg.region = *value
		return nil
	}
}

func AccessKeyIDPtr(value *string) Option {
	return func(cfg *config) error {
		cfg.accessKeyID = *value
		return nil
	}
}

func SecretAccessKeyPtr(value *string) Option {
	return func(cfg *config) error {
		cfg.secretAccessKey = *value
		return nil
	}
}

func TokenPtr(value *string) Option {
	return func(cfg *config) error {
		cfg.secretToken = *value
		return nil
	}
}

func BucketPtr(value *string) Option {
	return func(cfg *config) error {
		cfg.bucket = *value
		return nil
	}
}

// EndpointFromEnv reads the endpoint in the following order,
// if not set:
//
// - DBFS_MINIO_ENDPOINT
// - DBFS_BUCKET_ENDPOINT
//
// If nothing is found, then it assumes localhost:9000
func EndpointFromEnv(cfg *config) error {
	if len(cfg.endpoint) > 0 {
		return nil
	}
	cfg.endpoint, _ = firstNonEmptyEnv("DBFS_MINIO_ENDPOINT", "DBFS_BUCKET_ENDPOINT")
	if cfg.endpoint == "" {
		// default minio location
		cfg.endpoint = "localhost:9000"
	}
	return nil
}

// AccesKeyIDFromEnv returns the key id in the following order,
// if not set:
//
// DBFS_MINIO_ACCESS_KEY_ID
// DBFS_BUCKET_ACCESS_KEY_ID
// AWS_ACCESS_KEY_ID
// MINIO_USER
// MINIO_ROOT_USER
//
// If nothing is found, an error is returned
func AccessKeyIDFromEnv(cfg *config) error {
	if len(cfg.accessKeyID) > 0 {
		return nil
	}
	var err error
	cfg.accessKeyID, err = firstNonEmptyEnv("DBFS_MINIO_ACCESS_KEY_ID", "DBFS_BUCKET_ACCESS_KEY_ID", "AWS_ACCESS_KEY_ID",
		"MINIO_USER", "MINIO_ROOT_USER")
	return err
}

// SecretAccessKeyFromEnv returns the key id in the following order,
// if not set:
//
// DBFS_MINIO_SECRET_ACCESS_KEY
// DBFS_BUCKET_SECRET_ACCESS_KEY
// AWS_SECRET_ACCESS_KEY
// MINIO_PASSWORD
// MINIO_ROOT_PASSWORD
//
// If nothing is found, an error is returned
func SecretAccessKeyFromEnv(cfg *config) error {
	if len(cfg.secretAccessKey) > 0 {
		return nil
	}
	var err error
	cfg.secretAccessKey, err = firstNonEmptyEnv("DBFS_MINIO_SECRET_ACCESS_KEY", "DBFS_BUCKET_SECRET_ACCESS_KEY",
		"AWS_SECRET_ACCESS_KEY", "MINIO_PASSWORD", "MINIO_ROOT_PASSWORD")
	return err
}

// TokenFromEnv returns the key id in the following order,
// if not set:
//
// DBFS_MINIO_SESSION_TOKEN
// DBFS_BUCKET_SESSION_TOKEN
// AWS_SESSION_TOKEN
// MINIO_TOKEN
//
// If nothing is found, an error is returned
func TokenFromEnv(cfg *config) error {
	if len(cfg.secretToken) > 0 {
		return nil
	}
	cfg.secretToken, _ = firstNonEmptyEnv("DBFS_MINIO_SESSION_TOKEN", "DBFS_BUCKET_SESSION_TOKEN",
		"AWS_SESSION_TOKEN", "MINIO_TOKEN")
	return nil
}

// BucketFromEnv returns the bucket in the following order,
// if not set:
// DBFS_MINIO_BUCKET_NAME
// DBFS_BUCKET_NAME
//
// If nothing is set, the name `dbfs` is assumed
func BucketFromEnv(cfg *config) error {
	if len(cfg.bucket) > 0 {
		return nil
	}
	cfg.bucket, _ = firstNonEmptyEnv("DBFS_MINIO_BUCKET_NAME", "DBFS_BUCKET_NAME", "DBFS_MINIO_BUCKET", "DBFS_BUCKET")
	if len(cfg.bucket) == 0 {
		cfg.bucket = "dbfs"
	}
	return nil
}

// RegionFromEnv returns the region in the following order,
// if not set:
// DBFS_MINIO_BUCKET_REGION
// DBFS_BUCKET_REGION
// AWS_DEFAULT_REGION
//
// If nothing is set, an error is returned
func RegionFromEnv(cfg *config) error {
	if len(cfg.region) > 0 {
		return nil
	}
	var err error
	cfg.region, err = firstNonEmptyEnv("DBFS_MINIO_BUCKET_REGION", "DBFS_BUCKET_REGION", "AWS_DEFAULT_REGION")
	return err
}

func firstNonEmptyEnv(names ...string) (string, error) {
	for _, n := range names {
		v := os.Getenv(n)
		if len(v) > 0 {
			return v, nil
		}
	}
	return "", fmt.Errorf("one of these variables [%v] must be defined", names)
}

// Connect to a minio host, if options is empty,
// then FromEnv is used instead
func Connect(ctx context.Context, options ...Option) (*Bucket, error) {
	cfg := config{}
	if len(options) == 0 {
		options = FromEnv()
	}
	for _, opt := range options {
		err := opt(&cfg)
		if err != nil {
			return nil, err
		}
	}
	minioCli, err := minio.New(cfg.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.accessKeyID, cfg.secretAccessKey, cfg.secretToken),
		Region: cfg.region,
	})
	if err != nil {
		return nil, err
	}
	return &Bucket{cli: minioCli, bucket: cfg.bucket, region: cfg.region}, nil
}

func (b *Bucket) Close() error { return nil }
func (b *Bucket) Copy(ctx context.Context, to, from string) error {
	_, err := b.cli.CopyObject(ctx, minio.CopyDestOptions{
		Bucket: b.bucket,
		Object: to,
	}, minio.CopySrcOptions{
		Bucket: b.bucket,
		Object: from,
	})
	return err
}
func (b *Bucket) Delete(ctx context.Context, from string) error {
	return b.cli.RemoveObject(ctx, b.bucket, from, minio.RemoveObjectOptions{})
}
func (b *Bucket) Read(ctx context.Context, w io.Writer, from string) (int64, error) {
	obj, err := b.cli.GetObject(ctx, b.bucket, from, minio.GetObjectOptions{})
	if err != nil {
		return 0, err
	}
	defer obj.Close()
	return io.Copy(w, obj)
}
func (b *Bucket) Write(ctx context.Context, to string, r io.Reader) (int64, error) {
	cr := countReader{actual: r}
	_, err := b.cli.PutObject(ctx, b.bucket, to, &cr, -1, minio.PutObjectOptions{})
	if err != nil {
		return 0, err
	}
	return cr.total, err
}
func (b *Bucket) Exists(ctx context.Context, path string) (bool, error) {
	stat, err := b.cli.StatObject(ctx, b.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return false, err
	}
	return !stat.IsDeleteMarker && stat.Size > 0 && stat.Err == nil, nil
}

func (cr *countReader) Read(b []byte) (int, error) {
	n, e := cr.actual.Read(b)
	cr.total += int64(n)
	return n, e
}
