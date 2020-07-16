package authfs

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestRegister(t *testing.T) {
	afero.NewOsFs().MkdirAll(filepath.FromSlash("./testdata"), 0544)
	fs := afero.NewBasePathFs(afero.NewOsFs(), filepath.Join("./testdata"))
	c := Open(fs)
	if err := c.Register("user", []byte("user")); err != nil {
		t.Fatal(err)
	}
	if authenticated, err := c.Authenticate("user", []byte("user")); err != nil {
		t.Fatal(err)
	} else if !authenticated {
		t.Fatal("should have been authenticated")
	}
}
