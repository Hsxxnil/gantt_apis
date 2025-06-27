package task

import (
	present "gantt/internal/presenter/task"
	"gantt/internal/router/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("gantt").Group("v1.0").Group("tasks")
	{
		v10.POST("", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Create)
		v10.POST("create-all", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.CreateAll)
		v10.POST("import", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Import)
		v10.POST("get-by-projects", middleware.Verify(), middleware.CheckPermission(), control.GetByProjectUUIDList)
		v10.GET(":taskUUID", middleware.Verify(), middleware.CheckPermission(), control.GetBySingle)
		v10.GET("no-pagination/no-sub-filter", middleware.Verify(), middleware.CheckPermission(), control.GetByListNoPaginationNoSub)
		v10.DELETE("", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Delete)
		v10.PATCH(":taskUUID", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Update)
		v10.PATCH("update-all", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.UpdateAll)
	}

	return router
}
