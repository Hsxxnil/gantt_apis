package main

import (
	"gantt/internal/interactor/pkg/connect"
	"gantt/internal/interactor/pkg/util/log"
	"gantt/internal/router"
	"gantt/internal/router/event_mark"
	"gantt/internal/router/project"
	"gantt/internal/router/project_resource"
	"gantt/internal/router/project_type"

	"github.com/apex/gateway"
)

func main() {
	db, err := connect.PostgresSQL()
	if err != nil {
		log.Error(err)
		return
	}

	engine := router.Default()
	engine = project.GetRouter(engine, db)
	engine = project_type.GetRouter(engine, db)
	engine = project_resource.GetRouter(engine, db)
	engine = event_mark.GetRouter(engine, db)

	log.Fatal(gateway.ListenAndServe(":8080", engine))
}
