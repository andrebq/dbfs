// pakcage cas provides a content-addressable storage interface
// on top of an gocloud compatible interface.
//
// The implementation is written in such way that any service
// that provides an object storage (s3-like apis) can be used.
//
// MinIO is recommended for instalations that don't rely on
// a cloud provider
package cas
