package main

import (
	"gantt/internal/interactor/pkg/connect"
	"gantt/internal/interactor/pkg/util/log"
	"gantt/internal/router"
	"gantt/internal/router/holiday"
	"gantt/internal/router/work_day"

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
