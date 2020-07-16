package webdav

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

type (
	dir struct {
		data afero.Fs
	}
)

func newDir(data afero.Fs) webdav.FileSystem {
	return &dir{
		data: data,
	}
}

func (d *dir) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return d.data.Mkdir(filepath.FromSlash(name), perm)
}

func (d *dir) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return d.data.OpenFile(filepath.FromSlash(name), flag, perm)
}

func (d *dir) RemoveAll(ctx context.Context, name string) error {
	return d.data.RemoveAll(filepath.FromSlash(name))
}

func (d *dir) Rename(ctx context.Context, oldName, newName string) error {
	return d.data.Rename(filepath.FromSlash(oldName), filepath.FromSlash(newName))
}

func (d *dir) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return d.data.Stat(filepath.FromSlash(name))
}
