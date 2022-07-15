package postgres

import (
	_ "database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"os"
)

type Connector struct {
	DB *sqlx.DB
}

func (p *Connector) Connect() {
	connInfo := os.Getenv("APP_ORDERS_SERVICE_DB")
	if connInfo == "" {
		log.Fatal().Msg("empty postgres host")
	}

	var err error
	p.DB, err = sqlx.Open("postgres", connInfo)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to postgres")
	}

	p.DB.SetMaxOpenConns(50)
	p.DB.SetMaxIdleConns(0)
	p.DB.SetConnMaxLifetime(0)
	err = p.DB.Ping()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to postgres")
	}
	log.Info().Str("DSN", connInfo).Msg("connected to postgres")
}
