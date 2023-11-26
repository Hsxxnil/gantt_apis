package task

import (
	present "hta/internal/presenter/task"
	"hta/internal/router/middleware"
	"hta/internal/router/middleware/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("hta-gantt").Group("v1.0").Group("tasks")
	{
		v10.POST("", middleware.Verify(), auth.CheckPermission(), middleware.Transaction(db), control.Create)
		v10.POST("create-all", middleware.Verify(), auth.CheckPermission(), middleware.Transaction(db), control.CreateAll)
		v10.POST("import", middleware.Verify(), auth.CheckPermission(), middleware.Transaction(db), control.Import)
		v10.POST("get-by-projects", middleware.Verify(), auth.CheckPermission(), control.GetByProjectUUIDList)
		v10.GET(":taskUUID", middleware.Verify(), auth.CheckPermission(), control.GetBySingle)
		v10.GET("no-pagination/no-sub-filter", middleware.Verify(), auth.CheckPermission(), control.GetByListNoPaginationNoSub)
		v10.DELETE("", middleware.Verify(), auth.CheckPermission(), middleware.Transaction(db), control.Delete)
		v10.PATCH(":taskUUID", middleware.Verify(), auth.CheckPermission(), middleware.Transaction(db), control.Update)
		v10.PATCH("update-all", middleware.Verify(), auth.CheckPermission(), middleware.Transaction(db), control.UpdateAll)
	}

	return router
}
