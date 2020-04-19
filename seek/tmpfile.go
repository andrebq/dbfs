package seek

import (
	"io"
	"io/ioutil"
	"os"
)

type (
	// ReadSeekCloser adds Closer to ReadSeeker
	ReadSeekCloser interface {
		io.ReadSeeker
		io.Closer
	}

	tmpFile struct {
		original io.Reader
		seeker   io.Seeker
		reader   io.Reader
		closer   io.Closer
		init     bool
	}

	closeNRemove struct {
		f *os.File
	}

	errSeeker struct{ err error }
	errReader struct{ err error }
	errCloser struct{ err error }
)

func (e errSeeker) Seek(_ int64, _ int) (int64, error) { return 0, e.err }
func (e errReader) Read(_ []byte) (int, error)         { return 0, e.err }
func (e errCloser) Close() error                       { return e.err }
func (c closeNRemove) Close() error {
	f := c.f
	if f == nil {
		return os.ErrClosed
	}

	defer os.Remove(f.Name())
	return f.Close()
}

// Read implements io.Reader
func (t tmpFile) Read(r []byte) (int, error) {
	if !t.init {
		t.setup()
	}
	return t.reader.Read(r)
}

//  Seek implements io.Seeker
func (t tmpFile) Seek(offset int64, whence int) (int64, error) {
	if !t.init {
		t.setup()
	}
	return t.seeker.Seek(offset, whence)
}

// Close implements io.Closer
func (t tmpFile) Close() error {
	// we did not even initializer it
	// no need to initialize just to close it
	if !t.init {
		return nil
	}
	return t.closer.Close()
}

func (t *tmpFile) setup() {
	if t.init {
		return
	}
	t.init = true
	tmp, err := ioutil.TempFile("", "-seeker.tmp")
	if err != nil {
		t.closer = errCloser{err}
		t.seeker = errSeeker{err}
		t.reader = errReader{err}
		return
	}
	t.closer = closeNRemove{tmp}
	t.seeker = tmp
	t.reader = tmp
	_, err = io.Copy(tmp, t.original)
	if err != nil {
		n := tmp.Name()
		tmp.Close()
		os.Remove(n)
		t.closer = errCloser{err}
		t.seeker = errSeeker{err}
		t.reader = errReader{err}
	}
	_, err = t.seeker.Seek(0, 0)
	if err != nil {
		n := tmp.Name()
		tmp.Close()
		os.Remove(n)
		t.closer = errCloser{err}
		t.seeker = errSeeker{err}
		t.reader = errReader{err}
	}
}

// CopyToTemp create a temporary file and lazily copies input to it
//
// When closed, the temporary file is deleted, so it is adivised to call Close
// in order to reclaim space earlier
func CopyToTemp(r io.Reader) ReadSeekCloser {
	return tmpFile{original: r}
}
