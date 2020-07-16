// Package authfs provides authentication based on files in the filesystem
package authfs

import (
	"errors"
	"fmt"
	"path"

	"github.com/andrebq/dbfs/internal/cborfs"
	"github.com/asaskevich/govalidator"
	"github.com/spf13/afero"
	"golang.org/x/crypto/bcrypt"
)

type (
	// Catalog contains the information about all users
	Catalog struct {
		root afero.Fs
	}

	// Identity contains information from an entry in the
	// catalog
	Identity struct {
		Name   string `cbor:",omitempty"`
		Hashed []byte `cbor:",omitempty"`
	}
)

// Open a catalog of users
func Open(fs afero.Fs) *Catalog {
	return &Catalog{root: fs}
}

// Authenticate returns true only, and only if, the username and password
// pair are correct
func (c *Catalog) Authenticate(name string, password []byte) (bool, error) {
	if !govalidator.IsASCII(name) {
		// TODO(andre): lift this restriction to allow for anything that is
		// a valid email address (without host)
		return false, errors.New("name must be a valid ASCII character")
	}
	if len(name) == 0 || len(password) == 0 {
		return false, nil
	}
	var id Identity
	err := cborfs.ReadFile(&id, c.root, path.Join("catalog", "users", name))
	if err != nil {
		return false, err
	}
	if id.Name != name {
		return false, nil
	}

	err = bcrypt.CompareHashAndPassword(id.Hashed, password)
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// Register adds the given username into the catalog
func (c *Catalog) Register(name string, password []byte) error {
	if !govalidator.IsASCII(name) {
		// TODO(andre): lift this restriction to allow for anything that is
		// a valid email address (without host)
		return errors.New("name must be a valid ASCII character")
	}
	if len(name) == 0 || len(password) == 0 {
		return errors.New("missing username and password")
	}
	var id Identity
	id.Name = name
	var err error
	id.Hashed, err = bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("unable to compute hashed password: %w", err)
	}
	err = cborfs.Overwrite(c.root, path.Join("catalog", "users", name), 0600, &id)
	if err != nil {
		fmt.Errorf("unable to save user: %v", err)
	}
	return nil
}
