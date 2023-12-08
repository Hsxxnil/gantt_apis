package main

import (
	"hta/internal/interactor/pkg/connect"
	"hta/internal/interactor/pkg/util/log"
	"hta/internal/router"
	"hta/internal/router/company"
	"hta/internal/router/event_mark"
	"hta/internal/router/holiday"
	"hta/internal/router/login"
	"hta/internal/router/policy"
	"hta/internal/router/project"
	"hta/internal/router/project_resource"
	"hta/internal/router/project_type"
	"hta/internal/router/resource"
	"hta/internal/router/role"
	"hta/internal/router/task"
	"hta/internal/router/user"
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
	engine = resource.GetRouter(engine, db)
	engine = task.GetRouter(engine, db)
	engine = project.GetRouter(engine, db)
	engine = holiday.GetRouter(engine, db)
	engine = event_mark.GetRouter(engine, db)
	engine = work_day.GetRouter(engine, db)
	engine = project_type.GetRouter(engine, db)
	engine = project_resource.GetRouter(engine, db)
	engine = user.GetRouter(engine, db)
	engine = login.GetRouter(engine, db)
	engine = policy.GetRouter(engine, db)
	engine = role.GetRouter(engine, db)
	engine = company.GetRouter(engine, db)

	log.Fatal(gateway.ListenAndServe(":8080", engine))
}
