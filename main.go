package main

import (
	"fmt"
	"hta/internal/interactor/pkg/connect"
	"hta/internal/router/department"
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
	"net/http"

	_ "hta/api"
	"hta/internal/interactor/pkg/util/log"
	"hta/internal/router"

	//"hta/internal/router/permission"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// main is run all api form localhost port 8080

//	@title			HTA GANTT APIs
//	@version		0.1
//	@description	HTA2 GANTT APIs
//	@termsOfService

//	@contact.name
//	@contact.url
//	@contact.email

//	@license.name	AGPL 3.0
//	@license.url	https://www.gnu.org/licenses/agpl-3.0.en.html

// @host		pmip.t.api.likbox.com
// @BasePath	/hta-gantt/v1.0
// @schemes	https
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

	url := ginSwagger.URL(fmt.Sprintf("http://localhost:8080/swagger/doc.json"))
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	log.Fatal(http.ListenAndServe(":18080", engine))
}
