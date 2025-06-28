package main

import (
	"fmt"
	"gantt/internal/interactor/pkg/connect"
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
	"net/http"

	_ "gantt/api"
	"gantt/internal/interactor/pkg/util/log"
	"gantt/internal/router"

	//"gantt/internal/router/permission"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// main is run all api form localhost port 8080

//	@title			GANTT APIs
//	@version		0.1
//	@description	GANTT APIs
//	@termsOfService

//	@contact.name
//	@contact.url
//	@contact.email

//	@license.name	AGPL 3.0
//	@license.url	https://www.gnu.org/licenses/agpl-3.0.en.html

// @host		localhost:18080
// @BasePath	/gantt/v1.0
// @schemes	http
func main() {
	db, err := connect.PostgresSQL()
	if err != nil {
		log.Error(err)
		return
	}

	engine := router.Default()
	resource.GetRouter(engine, db)
	project.GetRouter(engine, db)
	holiday.GetRouter(engine, db)
	event_mark.GetRouter(engine, db)
	work_day.GetRouter(engine, db)
	project_type.GetRouter(engine, db)
	project_resource.GetRouter(engine, db)
	user.GetRouter(engine, db)
	login.GetRouter(engine, db)
	policy.GetRouter(engine, db)
	role.GetRouter(engine, db)
	task.GetRouter(engine, db)
	department.GetRouter(engine, db)
	s3_file.GetRouter(engine, db)

	url := ginSwagger.URL(fmt.Sprintf("http://localhost:8080/swagger/doc.json"))
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	log.Fatal(http.ListenAndServe(":18080", engine))
}
