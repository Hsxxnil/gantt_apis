package resource

import (
	present "hta/internal/presenter/resource"
	"hta/internal/router/middleware"
	"hta/internal/router/middleware/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("hta-gantt").Group("v1.0").Group("resources")
	{
		v10.POST("", middleware.Verify(), auth.CheckPermission(), middleware.Transaction(db), control.Create)
		v10.POST("import", middleware.Verify(), auth.CheckPermission(), middleware.Transaction(db), control.Import)
		v10.POST("/list", middleware.Verify(), auth.CheckPermission(), control.GetByList)
		v10.GET("no-pagination", middleware.Verify(), auth.CheckPermission(), control.GetByListNoPagination)
		v10.GET(":resourceUUID", middleware.Verify(), auth.CheckPermission(), control.GetBySingle)
		v10.DELETE(":resourceUUID", middleware.Verify(), auth.CheckPermission(), control.Delete)
		v10.PATCH(":resourceUUID", middleware.Verify(), auth.CheckPermission(), control.Update)
	}

	return router
}
