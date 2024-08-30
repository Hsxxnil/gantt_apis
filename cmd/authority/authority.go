package main

import (
	"hta/internal/interactor/pkg/connect"
	"hta/internal/interactor/pkg/util/log"
	"hta/internal/router"
	"hta/internal/router/department"
	"hta/internal/router/login"
	"hta/internal/router/policy"
	"hta/internal/router/role"
	"hta/internal/router/user"

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
