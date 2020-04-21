package blob

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/spf13/afero"
)

type (
	// Ref is just an alias for a sha256 hash
	Ref [sha256.Size]byte

	// CAS represents a content addressable storage
	CAS interface {
		// Put io.Reader into the CAS and return the Ref
		Put(io.ReadSeeker) (Ref, error)

		// Copy writes to the given output the content of the provided Ref (if it exists)
		// or returns an error if the reference is invalid.
		Copy(io.Writer, Ref) error

		// List all references available
		List(output chan Ref, err chan error)
	}

	// FileCAS implements the CAS interface using a filesystem interface.
	FileCAS struct {
		fs afero.Fs
	}
)

var (
	_       CAS = FileCAS{}
	shapool     = sync.Pool{
		New: func() interface{} { return sha256.New() },
	}
)

// NewFileCAS returns an empty case using the given Filesystem and the provided root path
func NewFileCAS(fs afero.Fs, base string) (*FileCAS, error) {
	target := filepath.Join(filepath.Clean(base), "blobs")

	err := fs.MkdirAll(target, 0755)
	if err != nil {
		return nil, err
	}

	bs := afero.NewBasePathFs(fs, target)
	return &FileCAS{fs: bs}, nil
}

// OpenFileCAS opens the given directory and within an isolated namespace
func OpenFileCAS(fp string, ns string) (*FileCAS, error) {
	var ofs afero.Fs
	fp = filepath.Clean(fp)
	if len(fp) > 0 && fp != "." {
		err := os.MkdirAll(filepath.Clean(fp), 0755)
		if err != nil {
			return nil, err
		}
		ofs = afero.NewBasePathFs(afero.NewOsFs(), fp)
	} else {
		ofs = afero.NewOsFs()
	}
	return NewFileCAS(ofs, ns)
}

// Copy implements CAS
func (f FileCAS) Copy(w io.Writer, r Ref) error {
	p := f.computePathOf(&r)
	exists, err := afero.Exists(f.fs, p)
	if err != nil {
		return fmt.Errorf("unable to find file: %w", err)
	}
	if !exists {
		return errNotFound
	}
	file, err := f.fs.OpenFile(p, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open file for read: %w", err)
	}
	defer file.Close()
	_, err = io.Copy(w, file)
	if err != nil {
		return fmt.Errorf("unable to copy contents: %w", err)
	}
	return nil
}

// Put computes the sha256 of the given content and moves it to the final location.
func (f FileCAS) Put(r io.ReadSeeker) (Ref, error) {
	h := shapool.Get().(hash.Hash)
	h.Reset()
	defer shapool.Put(h)
	_, err := io.Copy(h, r)
	if err != nil {
		return Ref{}, fmt.Errorf("unable to compute ref: %w", err)
	}
	_, err = r.Seek(0, io.SeekStart)
	if err != nil {
		return Ref{}, fmt.Errorf("unable to seek to the beginning of the reader: %v", err)
	}
	var ref Ref
	h.Sum(ref[:0])
	p := f.computePathOf(&ref) + ".draft"
	dir := path.Dir(p)
	err = f.fs.MkdirAll(dir, 0755)
	if err != nil {
		return Ref{}, fmt.Errorf("unable to create parent directory: %w", err)
	}
	target, err := f.fs.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		_, err := f.fs.Stat(f.computePathOf(&ref))
		if err == afero.ErrFileNotFound {
			// another process is working on the file we have to abort and wait
			return Ref{}, fmt.Errorf("unable to acquire exclusive access to the target file: %v", err)
		}
		// the file exists, so we do not need to copy it again
		return ref, nil
	}
	_, err = io.Copy(target, r)
	// should handle it properly, but for now (famous last words), let's live with this...
	defer target.Close()

	// reset the pointer
	r.Seek(0, 0)
	if err != nil {
		target.Close()
		return Ref{}, fmt.Errorf("unable to copy contents to the final destinatio: %v", err)
	}
	err = target.Sync()
	if err != nil {
		// all bets are off if you cannot sync, so the safest route
		// is to remove it altogether and let the caller retry
		err = f.fs.Remove(p)
		if err != nil {
			// ouch!!! we cannot even remove the file, something really weird is happening
			// let's make sure that something in the message points out that we might have a corrupted
			// database (not just one file)
			return Ref{}, fmt.Errorf("[FATAL] unable to remove corrupted file database in invalid state: %v", err)
		}
		return Ref{}, fmt.Errorf("unable to sync target file, operation is safe to retry: %v", err)
	}
	target.Close()
	err = f.fs.Rename(p, p[:len(p)-len(".draft")])
	if err != nil {
		return ref, fmt.Errorf("unable to move temporary file to final destination: %v", err)
	}
	return ref, nil
}

// List implements CAS interface
func (f FileCAS) List(out chan Ref, errCh chan error) {
	if out == nil {
		return
	}
	defer func() {
		close(out)
		if errCh != nil {
			close(errCh)
		}
	}()
	afero.Walk(f.fs, ".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if errCh != nil {
				errCh <- err
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		var r Ref
		if err := toRef(&r, filepath.Base(info.Name())); err != nil {
			return nil
		}
		out <- r
		return nil
	})
}

func (f *FileCAS) computePathOf(r *Ref) string {
	a, b, c, d := toHex(r[:1]), toHex(r[1:2]), toHex(r[2:3]), toHex(r[:])
	return path.Join(a, b, c, d)
}

func toHex(b []byte) string { return hex.EncodeToString(b) }
func toRef(r *Ref, h string) error {
	_, err := hex.Decode(r[:], []byte(h))
	return err
}
