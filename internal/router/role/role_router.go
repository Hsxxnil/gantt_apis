package role

import (
	present "hta/internal/presenter/role"
	"hta/internal/router/middleware"
	"hta/internal/router/middleware/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("hta-gantt").Group("v1.0").Group("roles")
	{
		v10.POST("", middleware.Verify(), auth.CheckPermission(), middleware.Transaction(db), control.Create)
		v10.GET("", middleware.Verify(), auth.CheckPermission(), control.GetByList)
		v10.GET(":id", middleware.Verify(), auth.CheckPermission(), control.GetBySingle)
		v10.DELETE(":id", middleware.Verify(), auth.CheckPermission(), control.Delete)
		v10.PATCH(":id", middleware.Verify(), auth.CheckPermission(), control.Update)
	}

	return router
}
