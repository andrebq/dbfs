package webdav

import (
	"context"
	"os"

	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

type (
	dir struct {
		users afero.Fs
		data  afero.Fs
	}
)

func newDir(data, auth afero.Fs) webdav.FileSystem {
	return &dir{
		users: auth,
		data:  data,
	}
}

func (d *dir) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return nil
}

func (d *dir) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return nil, nil
}

func (d *dir) RemoveAll(ctx context.Context, name string) error {
	return nil
}

func (d *dir) Rename(ctx context.Context, oldName, newName string) error {
	return nil
}

func (d *dir) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return nil, nil
}
