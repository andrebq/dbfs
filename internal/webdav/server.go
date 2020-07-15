// Package webdav provides as WebDAV compatible interface on top of dbfs
//
// TODO(andre): actually make use of dbfs, for now, lets just validate
// the ergonomics of exposing a webdav server
package webdav

import (
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

type (
	// Config contains the configuration for a WebDAV server
	Config struct {
		rootfs afero.Fs
		prefix string
	}

	server struct {
		srv *webdav.Handler

		authfs afero.Fs
		datafs afero.Fs
	}
)

var (
	// ErrMissingRootFS indicates that a configuration does not contain
	// a location where the files should be kept
	ErrMissingRootFS = errors.New("missing rootfs")
)

// WithRootFS sets the root location where all data will be kept
func (c Config) WithRootFS(fs afero.Fs) Config {
	c.rootfs = fs
	return c
}

// WithPrefix changes the perfix which is removed from each request
// useful to host this dbfs in the same domain as others.
func (c Config) WithPrefix(p string) Config {
	c.prefix = p
	return c
}

func (c *Config) apply(s *server) error {
	if c.rootfs == nil {
		return ErrMissingRootFS
	}
	s.authfs = afero.NewBasePathFs(c.rootfs, "auth")
	s.datafs = afero.NewBasePathFs(c.rootfs, "data")
	s.srv = &webdav.Handler{
		Prefix:     c.prefix,
		FileSystem: newDir(s.datafs, s.authfs),
		LockSystem: webdav.NewMemLS(),
		Logger:     s.logRequests,
	}
	return nil
}

// NewServer returns a new webdav server
func NewServer(c Config) (http.Handler, error) {
	s := &server{}
	err := c.apply(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if err := s.authorize(req); err != nil {
		s.renderErr(w, req, err)
	}
	s.srv.ServeHTTP(w, req)
}

func (s *server) authorize(req *http.Request) error {
	return errors.New("not implemented")
}

func (s *server) renderErr(w http.ResponseWriter, req *http.Request, err error) {
	// TODO(andre) add sampled log here as this is user driven
	log.Error().Err(err).Send()
}

func (s *server) logRequests(req *http.Request, err error) {
	// TODO(andre) think about how to log stuff here
}
