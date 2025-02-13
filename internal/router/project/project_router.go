package project

import (
	present "gantt/internal/presenter/project"
	"gantt/internal/router/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("hta-gantt").Group("v1.0").Group("projects")
	{
		v10.POST("", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Create)
		v10.POST("list", middleware.Verify(), middleware.CheckPermission(), control.GetByList)
		v10.GET("no-pagination", middleware.Verify(), middleware.CheckPermission(), control.GetByListNoPagination)
		v10.GET(":projectID", middleware.Verify(), middleware.CheckPermission(), control.GetBySingle)
		v10.DELETE(":projectID", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Delete)
		v10.PATCH(":projectID", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Update)
	}

	return router
}
