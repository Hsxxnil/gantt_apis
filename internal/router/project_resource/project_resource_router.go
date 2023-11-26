package project_resource

import (
	present "hta/internal/presenter/project_resource"
	"hta/internal/router/middleware"
	"hta/internal/router/middleware/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("hta-gantt").Group("v1.0").Group("project-resources")
	{
		v10.POST("", middleware.Verify(), auth.CheckPermission(), control.GetByList)
		v10.GET(":id", middleware.Verify(), auth.CheckPermission(), control.GetBySingle)
		v10.POST("get-by-project", middleware.Verify(), auth.CheckPermission(), control.GetByProjectList)
	}

	return router
}
