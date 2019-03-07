package environment

import (
	"github.com/OlegSchwann/rpsarena-ru-backend/authorization_server/accessor"
)

type Environment struct {
	DB     accessor.DB
	Config Config
}

type Config struct {
	PostgresPath  *string
	ListeningPort *string
	ImagesRoot    *string
}
