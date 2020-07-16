package main

import (
	"flag"

	"github.com/rs/zerolog/log"

	"github.com/andrebq/dbfs/internal/authfs"
	"github.com/spf13/afero"
)

var (
	user     = flag.String("user", "", "Username to add")
	password = flag.String("password", "", "Password to use")
	root     = flag.String("root", "", "Folder to keep the authentication information")
)

func main() {
	fs := afero.NewOsFs()
	c := authfs.Open(fs)

	err := c.Register(*user, []byte(*password))
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to add user")
	}
}
