package project_resource

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	present "hta/internal/presenter/project_resource"
	"hta/internal/router/middleware"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("hta-gantt").Group("v1.0").Group("project-resources")
	{
		v10.POST("", middleware.Verify(), middleware.CheckPermission(), control.GetByList)
		v10.GET(":id", middleware.Verify(), middleware.CheckPermission(), control.GetBySingle)
		v10.POST("get-by-project", middleware.Verify(), middleware.CheckPermission(), control.GetByProjectList)
	}

	return router
}
