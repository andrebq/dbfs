package cas

import (
	"context"
	"io"
)

type (
	// KV represents the bare minimum abstraction required by cas
	KV interface {
		io.Closer
		Write(context.Context, string, io.Reader) (int64, error)
		Read(context.Context, io.Writer, string) (int64, error)
		Copy(context.Context, string, string) error
		Delete(context.Context, string) error
		Exists(context.Context, string) (bool, error)
	}

	// MovableObjects contains the methods required by cas which
	// help reduce the load on the remote server
	Mover interface {
		Move(context.Context, string, string) error
	}
)

// move objects from a location to another, if kv implements the
// mover interface, that method is used.
//
// Otherwise this happens as a two step operation, first a copy is made
// than the original object is removed.
func move(ctx context.Context, kv KV, to, from string) error {
	if mover, ok := kv.(Mover); ok {
		return mover.Move(ctx, to, from)
	}
	err := kv.Copy(ctx, to, from)
	if err != nil {
		return err
	}
	return kv.Delete(ctx, from)
}
