package cborfs

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/fxamacker/cbor"
	"github.com/spf13/afero"
)

// ReadFile reads a file at the given location and decodes its
// content to the given structure.
//
// The system expects that a file was encoded using CBOR
func ReadFile(out interface{}, fs afero.Fs, file string) error {
	vfile, err := fs.OpenFile(filepath.FromSlash(path.Clean(file)), os.O_RDONLY, 0400)
	if err != nil {
		return fmt.Errorf("unable to open file %v for reading: %w", file, err)
	}
	defer vfile.Close()
	return cbor.NewDecoder(vfile).Decode(out)
}

// Overwrite the file (or create a new one if needed). mode is only used if the file
// is new, otherwise, it is ignored.
func Overwrite(fs afero.Fs, file string, mode os.FileMode, input interface{}) error {
	file = filepath.FromSlash(path.Clean(file))
	dir := filepath.Dir(file)
	err := fs.MkdirAll(dir, 0544)
	if err != nil {
		return fmt.Errorf("unable to create directory for %v: %w", file, err)
	}
	vfile, err := fs.OpenFile(filepath.FromSlash(path.Clean(file)), os.O_WRONLY|os.O_CREATE, mode)
	if err != nil {
		return fmt.Errorf("unable to open/crete file %v for writing (%v): %w", file, mode, err)
	}
	defer vfile.Close()
	err = cbor.NewEncoder(vfile, cbor.CTAP2EncOptions()).Encode(input)
	if err != nil {
		return err
	}
	err = vfile.Sync()
	if err != nil {
		return err
	}
	return nil
}
