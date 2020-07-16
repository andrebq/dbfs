package main

import (
	"flag"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"github.com/andrebq/dbfs/internal/webdav"
	"github.com/spf13/afero"
)

var (
	rootDir = flag.String("root", ".", "Root dir where all data is kept")
	prefix  = flag.String("prefix", "", "Prefix which should be removed when resolving paths")
	port    = flag.Uint("port", 8080, "Port to listen for incoming requests")
	addr    = flag.String("addr", "0.0.0.0", "Interface address to bind for incomming requests")
)

func main() {
	flag.Parse()
	cfg := webdav.Config{}.
		WithRootFS(afero.NewBasePathFs(afero.NewOsFs(), filepath.Clean(*rootDir))).
		WithPrefix(*prefix)
	srv, err := webdav.NewServer(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create new server instance")
	}

	listenAddr := fmt.Sprintf("%v:%d", *addr, *port)
	log.Info().
		Str("app", "webdav").
		Str("addr", listenAddr).
		Str("root", filepath.Clean(*rootDir)).
		Str("prefix", *prefix).
		Msg("Starting server")

	// TODO(andre): properly handle sigterm
	err = http.ListenAndServe(listenAddr, srv)
	if err != nil {
		log.Error().Err(err).Msg("Unable to start server")
	}
}
