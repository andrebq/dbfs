// Package webdav provides as WebDAV compatible interface on top of dbfs
//
// TODO(andre): actually make use of dbfs, for now, lets just validate
// the ergonomics of exposing a webdav server
package webdav

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/andrebq/dbfs/internal/authfs"
	"github.com/gosimple/slug"
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

	userPartition struct {
		server webdav.Handler
	}

	server struct {
		sync.Mutex
		auth   *authfs.Catalog
		datafs afero.Fs
		prefix string

		users map[string]*userPartition
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
	s.auth = authfs.Open(afero.NewBasePathFs(c.rootfs, "auth"))
	s.datafs = afero.NewBasePathFs(c.rootfs, "files")
	s.users = make(map[string]*userPartition)
	s.prefix = c.prefix
	return nil
}

// NewServer returns a new webdav server
func NewServer(c Config) (http.Handler, error) {
	s := &server{}
	err := c.apply(s)
	if err != nil {
		return nil, err
	}
	return http.StripPrefix(c.prefix, s), nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var up *userPartition
	if user, err := s.authorize(req); err != nil {
		s.renderErr(w, req, err)
		return
	} else {
		s.Lock()
		up = s.users[user]
		if up == nil {
			up = &userPartition{
				server: webdav.Handler{
					FileSystem: newDir(afero.NewBasePathFs(s.datafs, path.Clean(slug.Make(user)))),
					LockSystem: webdav.NewMemLS(),
					Logger:     s.logRequests,
				},
			}
			s.users[user] = up
		}
		s.Unlock()
	}
	up.server.ServeHTTP(w, req)
}

func (s *server) authorize(req *http.Request) (string, error) {
	user, pwd, contains := req.BasicAuth()
	if !contains {
		return "", authRequired{realm: fmt.Sprintf("%v-webdav", req.Host)}
	}
	ok, err := s.auth.Authenticate(user, []byte(pwd))
	if err != nil {
		log.Error().Err(err).Msg("error authenticating request")
		return "", httpError{
			cause:  err,
			msg:    "internal error",
			status: http.StatusInternalServerError,
		}
	}
	if !ok {
		return "", authRequired{realm: fmt.Sprintf("%v-webdav", req.Host)}
	}
	return user, nil
}

func (s *server) renderErr(w http.ResponseWriter, req *http.Request, err error) {
	// TODO(andre) add sampled log here as this is user driven
	tick := time.Now().Unix() ^ rand.Int63()
	log.Error().Int64("err-token", tick).Err(err).Send()
	if render, ok := err.(interface {
		Render(http.ResponseWriter, *http.Request)
	}); ok {
		render.Render(w, req)
		return
	}

	http.Error(w, fmt.Sprintf("Unexpected error (error token: %v)", tick), http.StatusInternalServerError)
}

func (s *server) logRequests(req *http.Request, err error) {
	// TODO(andre) think about how to log stuff here
}
