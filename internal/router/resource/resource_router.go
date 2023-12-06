package resource

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	present "hta/internal/presenter/resource"
	"hta/internal/router/middleware"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("hta-gantt").Group("v1.0").Group("resources")
	{
		v10.POST("", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Create)
		v10.POST("import", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Import)
		v10.POST("/list", middleware.Verify(), middleware.CheckPermission(), control.GetByList)
		v10.GET("no-pagination", middleware.Verify(), middleware.CheckPermission(), control.GetByListNoPagination)
		v10.GET(":resourceUUID", middleware.Verify(), middleware.CheckPermission(), control.GetBySingle)
		v10.DELETE(":resourceUUID", middleware.Verify(), middleware.CheckPermission(), control.Delete)
		v10.PATCH(":resourceUUID", middleware.Verify(), middleware.CheckPermission(), control.Update)
	}

	return router
}
