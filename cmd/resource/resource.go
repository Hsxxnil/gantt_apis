package main

import (
	"gantt/internal/interactor/pkg/connect"
	"gantt/internal/interactor/pkg/util/log"
	"gantt/internal/router"
	"gantt/internal/router/resource"

	"github.com/apex/gateway"
)

func main() {
	db, err := connect.PostgresSQL()
	if err != nil {
		log.Error(err)
		return
	}

	engine := router.Default()
	engine = resource.GetRouter(engine, db)

	log.Fatal(gateway.ListenAndServe(":8080", engine))
}
