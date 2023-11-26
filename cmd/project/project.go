package main

import (
	"github.com/apex/gateway"
	"hta/internal/interactor/pkg/connect"
	"hta/internal/interactor/pkg/util/log"
	"hta/internal/router"
	"hta/internal/router/event_mark"
	"hta/internal/router/project"
	"hta/internal/router/project_resource"
	"hta/internal/router/project_type"
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
