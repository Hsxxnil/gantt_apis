package postgres

import (
	"database/sql"
	"time"

	"gantt/internal/interactor/pkg/util/log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

type Config struct {
	// Customize Driver
	DriverName *string
	// Data Source Name
	DSN *string
	// DB connect pool interface
	Conn *sql.DB
	// Disables implicit prepared statement usage
	PreferSimpleProtocol *bool
	// Creates a prepared statement when executing any SQL and caches them to speed up future calls
	PrepareStmt *bool
	// Allow to change GORM’s default logger by overriding this option
	Logger logger.Interface
	// Change the function to be used when creating a new timestamp
	NowFunc func() time.Time
	// DBResolver adds multiple databases support
	Replicas []*string
}

func (c *Config) Connect() (db *gorm.DB, err error) {
	postgresConfig := postgres.Config{}
	gormConfig := gorm.Config{}

	if c.DSN != nil {
		postgresConfig.DSN = *c.DSN
	}

	if c.DriverName != nil {
		postgresConfig.DriverName = *c.DriverName
	}

	if c.Conn != nil {
		postgresConfig.Conn = c.Conn
	}

	if c.PreferSimpleProtocol != nil {
		postgresConfig.PreferSimpleProtocol = *c.PreferSimpleProtocol
	}

	if c.PrepareStmt != nil {
		gormConfig.PrepareStmt = *c.PrepareStmt
	}

	if c.Logger != nil {
		gormConfig.Logger = c.Logger
	}

	if c.NowFunc != nil {
		gormConfig.NowFunc = c.NowFunc
	}

	db, err = gorm.Open(postgres.New(postgresConfig), &gormConfig)
	if err != nil {
		return nil, err
	}

	var dialectics []gorm.Dialector
	for _, replica := range c.Replicas {
		director := postgres.New(postgres.Config{
			DSN:                  *replica,
			PreferSimpleProtocol: true,
		})
		dialectics = append(dialectics, director)
	}

	if c.Replicas != nil {
		err = db.Use(dbresolver.Register(dbresolver.Config{Replicas: dialectics}))
		if err != nil {
			log.Error(err)
		}
	}

	return db, nil
}
