package config

import "github.com/urfave/cli/v2"

type (
	// Storage configures the
	Storage struct {
		Driver string
		Minio  struct {
			Endpoint string
			Username string
			Password string
			Token    string
			Bucket   string
		}
	}
)

func (s *Storage) AllFlags() []cli.Flag {
	return []cli.Flag{
		s.DriverFlag(),
		s.MinioEndpointFlag(),
		s.MinioUsernameFlag(),
		s.MinioPasswordFlag(),
		s.MinioSessionTokenFlag(),
	}
}

func (s *Storage) DriverFlag() cli.Flag {
	return &cli.StringFlag{
		Hidden:      true,
		Name:        "storage-driver",
		EnvVars:     []string{"DBFS_STORAGE_DRIVER"},
		Usage:       "Driver to use for storage, only minio is supported",
		Value:       "minio",
		Destination: &s.Driver,
	}
}

func (s *Storage) MinioEndpointFlag() cli.Flag {
	return &cli.StringFlag{
		Name:        "minio-endpoint",
		EnvVars:     []string{"DBFS_MINIO_ENDPOINT", "DBFS_BUCKET_ENDPOINT"},
		Usage:       "Endpoint with minio content",
		Value:       "localhost:9000",
		Destination: &s.Minio.Endpoint,
	}
}

func (s *Storage) MinioUsernameFlag() cli.Flag {
	return &cli.StringFlag{
		Name: "minio-username",
		EnvVars: []string{"DBFS_MINIO_ACCESS_KEY_ID", "DBFS_BUCKET_ACCESS_KEY_ID", "AWS_ACCESS_KEY_ID",
			"MINIO_USER", "MINIO_ROOT_USER"},
		Usage:       "Username to authenticante against the minio server",
		Destination: &s.Minio.Username,
	}
}

func (s *Storage) MinioPasswordFlag() cli.Flag {
	return &cli.StringFlag{
		Name: "minio-password",
		EnvVars: []string{"DBFS_MINIO_SECRET_ACCESS_KEY", "DBFS_BUCKET_SECRET_ACCESS_KEY",
			"AWS_SECRET_ACCESS_KEY", "MINIO_PASSWORD", "MINIO_ROOT_PASSWORD"},
		Usage:       "Password to authenticante against the minio server",
		Destination: &s.Minio.Password,
	}
}

func (s *Storage) MinioSessionTokenFlag() cli.Flag {
	return &cli.StringFlag{
		Hidden: true,
		Name:   "minio-session-token",
		EnvVars: []string{"DBFS_MINIO_SESSION_TOKEN", "DBFS_BUCKET_SESSION_TOKEN",
			"AWS_SESSION_TOKEN", "MINIO_TOKEN"},
		Usage:       "Token to authenticante against the minio server",
		Destination: &s.Minio.Username,
	}
}
