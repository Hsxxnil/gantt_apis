package user

import (
	present "gantt/internal/presenter/user"
	"gantt/internal/router/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("hta-gantt").Group("v1.0").Group("users")
	{
		v10.POST("list", middleware.Verify(), middleware.CheckPermission(), control.GetByList)
		v10.POST("check-duplicate", control.Duplicate)
		v10.POST("authenticator/current-user", middleware.Verify(), control.EnableAuthenticator)
		v10.POST("reset-password/current-user", middleware.Verify(), control.ResetPassword)
		v10.POST("enable/current-user", middleware.Verify(), control.EnableByCurrent)
		v10.POST("change-email/current-user", middleware.Verify(), control.ChangeEmail)
		v10.POST("verify-email/current-user", middleware.Verify(), middleware.Transaction(db), control.VerifyEmail)
		v10.GET("", middleware.Verify(), middleware.CheckPermission(), control.GetByListNoPagination)
		v10.GET("current-user", middleware.Verify(), middleware.CheckPermission(), control.GetByCurrent)
		v10.GET(":id", middleware.Verify(), middleware.CheckPermission(), control.GetBySingle)
		v10.DELETE(":id", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Delete)
		v10.PATCH("current-user", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.UpdateByCurrent)
		v10.PATCH(":id", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Update)
	}

	return router
}
