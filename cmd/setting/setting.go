package main

import (
	"hta/internal/interactor/pkg/connect"
	"hta/internal/interactor/pkg/util/log"
	"hta/internal/router"
	"hta/internal/router/holiday"
	"hta/internal/router/work_day"

	"github.com/apex/gateway"
)

func main() {
	db, err := connect.PostgresSQL()
	if err != nil {
		log.Error(err)
		return
	}

	engine := router.Default()
	engine = holiday.GetRouter(engine, db)
	engine = work_day.GetRouter(engine, db)

	log.Fatal(gateway.ListenAndServe(":8080", engine))
}
