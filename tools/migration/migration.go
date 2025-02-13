package main

import (
	"fmt"
	"time"

	"gantt/config"
	dbConfig "gantt/internal/interactor/pkg/connect/postgres"
	"gantt/internal/interactor/pkg/util"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"gorm.io/gorm/logger"
)

func main() {
	const dsn string = "host=%s port=%d user=%s dbname=%s sslmode=%s password=%s"
	pgConfig := dbConfig.Config{}
	pgConfig.DSN = util.PointerString(
		fmt.Sprintf(dsn, config.SourceHost, config.SourcePort, config.SourceUser, config.SourceDataBase,
			config.SourceSSLMode, config.SourcePassword))
	pgConfig.PreferSimpleProtocol = util.PointerBool(true)
	pgConfig.NowFunc = func() time.Time { return time.Now().UTC() }
	pgConfig.Logger = logger.Default.LogMode(logger.Info)
	db, err := pgConfig.Connect()
	sourceDB, err := db.DB()
	if err != nil {
		fmt.Println(err)
	}

	driver, err := postgres.WithInstance(sourceDB, &postgres.Config{})
	if err != nil {
		fmt.Println(err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://./migrations", "postgres", driver)
	if err != nil {
		fmt.Println(err)
	}

	if err = m.Up(); err != nil {
		fmt.Println(err)
	}
}

// migrate create -ext sql -dir ./migrations create_xxx
// migrate create -ext sql -dir ./migrations alter_xxx
