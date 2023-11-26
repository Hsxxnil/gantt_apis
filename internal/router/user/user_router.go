package user

import (
	present "hta/internal/presenter/user"
	"hta/internal/router/middleware"
	"hta/internal/router/middleware/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("hta-gantt").Group("v1.0").Group("users")
	{
		v10.POST("", middleware.Verify(), middleware.Transaction(db), control.Create)
		v10.POST("list", middleware.Verify(), auth.CheckPermission(), control.GetByList)
		v10.GET("", middleware.Verify(), auth.CheckPermission(), control.GetByListNoPagination)
		v10.GET(":id", middleware.Verify(), auth.CheckPermission(), control.GetBySingle)
		v10.DELETE(":id", middleware.Verify(), auth.CheckPermission(), control.Delete)
		v10.PATCH(":id", middleware.Verify(), auth.CheckPermission(), control.Update)
	}

	return router
}
