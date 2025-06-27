package login

import (
	present "gantt/internal/presenter/login"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("gantt").Group("v1.0")
	{
		v10.POST("login", control.Login)
		v10.POST("verify", control.Verify)
		v10.POST("refresh", control.Refresh)
		v10.POST("forget-password", control.Forget)
		v10.POST("register", control.Register)
	}

	return router
}
