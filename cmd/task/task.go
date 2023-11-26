package main

import (
	"github.com/apex/gateway"
	"hta/internal/interactor/pkg/connect"
	"hta/internal/interactor/pkg/util/log"
	"hta/internal/router"
	"hta/internal/router/task"
)

func main() {
	db, err := connect.PostgresSQL()
	if err != nil {
		log.Error(err)
		return
	}

	engine := router.Default()
	engine = task.GetRouter(engine, db)

	log.Fatal(gateway.ListenAndServe(":8080", engine))
}
