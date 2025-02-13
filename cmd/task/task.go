package main

import (
	"gantt/internal/interactor/pkg/connect"
	"gantt/internal/interactor/pkg/util/log"
	"gantt/internal/router"
	"gantt/internal/router/s3_file"
	"gantt/internal/router/task"

	"github.com/apex/gateway"
)

func main() {
	db, err := connect.PostgresSQL()
	if err != nil {
		log.Error(err)
		return
	}

	engine := router.Default()
	engine = task.GetRouter(engine, db)
	engine = s3_file.GetRouter(engine, db)

	log.Fatal(gateway.ListenAndServe(":8080", engine))
}
