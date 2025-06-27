package role

import (
	present "gantt/internal/presenter/role"
	"gantt/internal/router/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("gantt").Group("v1.0").Group("roles")
	{
		v10.POST("", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Create)
		v10.GET("", middleware.Verify(), middleware.CheckPermission(), control.GetByList)
		v10.GET("no-pagination", middleware.Verify(), middleware.CheckPermission(), control.GetByListNoPagination)
		v10.GET(":id", middleware.Verify(), middleware.CheckPermission(), control.GetBySingle)
		v10.DELETE(":id", middleware.Verify(), middleware.CheckPermission(), control.Delete)
		v10.PATCH(":id", middleware.Verify(), middleware.CheckPermission(), control.Update)
	}

	return router
}
