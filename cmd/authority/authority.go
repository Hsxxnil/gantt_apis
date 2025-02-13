package main

import (
	"gantt/internal/interactor/pkg/connect"
	"gantt/internal/interactor/pkg/util/log"
	"gantt/internal/router"
	"gantt/internal/router/department"
	"gantt/internal/router/login"
	"gantt/internal/router/policy"
	"gantt/internal/router/role"
	"gantt/internal/router/user"

	"github.com/apex/gateway"
)

func main() {
	db, err := connect.PostgresSQL()
	if err != nil {
		log.Error(err)
		return
	}

	engine := router.Default()
	engine = user.GetRouter(engine, db)
	engine = login.GetRouter(engine, db)
	engine = policy.GetRouter(engine, db)
	engine = role.GetRouter(engine, db)
	engine = department.GetRouter(engine, db)

	log.Fatal(gateway.ListenAndServe(":8080", engine))
}
