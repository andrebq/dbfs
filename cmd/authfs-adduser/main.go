package main

import (
	"flag"

	"github.com/rs/zerolog/log"

	"github.com/andrebq/dbfs/internal/authfs"
	"github.com/gosimple/slug"
	"github.com/spf13/afero"
)

var (
	user     = flag.String("user", "", "Username to add")
	password = flag.String("password", "", "Password to use")
	root     = flag.String("root", "", "Folder to keep the authentication information")
	files    = flag.String("files", "", "If present, should point to a directory where the user home dir should be created")
)

func main() {
	flag.Parse()
	fs := afero.NewBasePathFs(afero.NewOsFs(), *root)
	c := authfs.Open(fs)

	err := c.Register(*user, []byte(*password))
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to add user")
	}

	if len(*files) != 0 {
		fs := afero.NewBasePathFs(afero.NewOsFs(), *files)
		err := fs.MkdirAll(slug.Make(*user), 0500)
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to create user home folder")
		}
	}
}
