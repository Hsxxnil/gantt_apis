package main

import (
	"gantt/internal/interactor/pkg/connect"
	"gantt/internal/interactor/pkg/util/log"
	"gantt/internal/router"
	"gantt/internal/router/department"
	"gantt/internal/router/event_mark"
	"gantt/internal/router/holiday"
	"gantt/internal/router/login"
	"gantt/internal/router/policy"
	"gantt/internal/router/project"
	"gantt/internal/router/project_resource"
	"gantt/internal/router/project_type"
	"gantt/internal/router/resource"
	"gantt/internal/router/role"
	"gantt/internal/router/s3_file"
	"gantt/internal/router/task"
	"gantt/internal/router/user"
	"gantt/internal/router/work_day"
	"os"
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
	engine = department.GetRouter(engine, db)
	engine = s3_file.GetRouter(engine, db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(gateway.ListenAndServe(":"+port, engine))
}
