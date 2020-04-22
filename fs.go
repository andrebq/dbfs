package dbfs

import (
	"io"
	"os"
	"path/filepath"

	"github.com/andrebq/dbfs/blob"
	"github.com/andrebq/dbfs/file"
	"github.com/andrebq/dbfs/seek"
	"github.com/spf13/afero"
)

type (
	// FilterFunc is used to filter files when running WriteDir
	//
	// Return true to allow processing of the given directory, false to skip it (including descendents)
	FilterFunc func(afero.Fs, string, os.FileInfo) (bool, error)
)

// CopyFile reads the content from r into cas using WriteFile under the hood
func CopyFile(cas blob.CAS, fs afero.Fs, path string) (blob.Ref, error) {
	f, err := fs.Open(path)
	if err != nil {
		return blob.Ref{}, err
	}
	defer f.Close()
	return WriteFile(cas, f)
}

// WriteFile copies r to f and storing all intermediate references into cas.
// the content of r is split into chunks of 1MB but if the file is too large,
// the file object might end up a value larger than 1MB
func WriteFile(cas blob.CAS, r io.Reader) (blob.Ref, error) {
	f := file.F{}
	m := file.Meta{
		Leaf: true,
		Size: 0,
	}
	buf := make([]byte, file.IdealChunkSize)
	eof := false
	for !eof {
		n, err := r.Read(buf)
		if err == io.EOF {
			eof = true
		} else if err != nil {
			return blob.Ref{}, err
		}
		if n == 0 {
			break
		}
		m.Size += int64(n)
		ref, err := cas.Put(seek.NewReadOnlyBuffer(buf[:n]))
		if err != nil {
			return blob.Ref{}, err
		}
		f.Chunks = append(f.Chunks, ref)
	}
	ref, err := cas.Put(seek.NewReadOnlyBuffer(m.ToBlob()))
	if err != nil {
		return blob.Ref{}, err
	}
	f.Meta = ref
	return cas.Put(seek.NewReadOnlyBuffer(f.ToBlob()))
}

// WriteDir opens folder and then writes each individual file/folder
// to cas (going recursively). A filter function can be used to decide
// if a directory/file should be stored or no.
func WriteDir(cas blob.CAS, fs afero.Fs, folder string, filter func(afero.Fs, string, os.FileInfo) (bool, error)) (blob.Ref, error) {
	d := file.F{}
	m := file.Meta{
		Leaf: false,
	}

	children, err := afero.ReadDir(fs, folder)
	if err != nil {
		return blob.Ref{}, err
	}

	for _, c := range children {
		valid, err := FilterFunc(filter).Call(fs, filepath.Join(folder, c.Name()), c)
		if err != nil {
			return blob.Ref{}, err
		}
		if !valid {
			continue
		}
		if c.IsDir() {
			ref, err := WriteDir(cas, fs, filepath.Join(folder, c.Name()), filter)
			if err != nil {
				return blob.Ref{}, err
			}
			d.Children = append(d.Children, file.NamedRef{
				Name: filepath.Base(c.Name()),
				Ref:  ref,
			})
		} else {
			ref, err := CopyFile(cas, fs, filepath.Join(folder, c.Name()))
			if err != nil {
				return blob.Ref{}, err
			}
			d.Children = append(d.Children, file.NamedRef{
				Name: filepath.Base(c.Name()),
				Ref:  ref,
			})
		}
	}

	ref, err := cas.Put(seek.NewReadOnlyBuffer(m.ToBlob()))
	if err != nil {
		return blob.Ref{}, err
	}
	d.Meta = ref
	return cas.Put(seek.NewReadOnlyBuffer(d.ToBlob()))
}

// Call runs the given filter
func (ff FilterFunc) Call(fs afero.Fs, path string, info os.FileInfo) (bool, error) {
	if ff == nil {
		return true, nil
	}
	return ff(fs, path, info)
}
