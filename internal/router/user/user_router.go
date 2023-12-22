package user

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	present "hta/internal/presenter/user"
	"hta/internal/router/middleware"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("hta-gantt").Group("v1.0").Group("users")
	{
		v10.POST("", middleware.Verify(), middleware.Transaction(db), control.Create)
		v10.POST("list", middleware.Verify(), middleware.CheckPermission(), control.GetByList)
		v10.GET("", middleware.Verify(), middleware.CheckPermission(), control.GetByListNoPagination)
		v10.GET("current-user", middleware.Verify(), middleware.CheckPermission(), control.GetByCurrent)
		v10.GET(":id", middleware.Verify(), middleware.CheckPermission(), control.GetBySingle)
		v10.DELETE(":id", middleware.Verify(), middleware.CheckPermission(), control.Delete)
		v10.PATCH("current-user", middleware.Verify(), middleware.CheckPermission(), control.Update)
		v10.PATCH("enable/:id", middleware.Verify(), middleware.CheckPermission(), control.Enable)
		v10.PATCH("enable/current-user", middleware.Verify(), middleware.CheckPermission(), control.EnableByCurrent)
		v10.PATCH("reset-password/current-user", middleware.Verify(), middleware.CheckPermission(), control.ResetPassword)
	}

	return router
}
