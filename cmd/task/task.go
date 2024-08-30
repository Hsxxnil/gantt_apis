package main

import (
	"hta/internal/interactor/pkg/connect"
	"hta/internal/interactor/pkg/util/log"
	"hta/internal/router"
	"hta/internal/router/s3_file"
	"hta/internal/router/task"

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
