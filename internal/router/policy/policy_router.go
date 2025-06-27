package policy

import (
	present "gantt/internal/presenter/policy"
	"gantt/internal/router/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init()
	v10 := router.Group("gantt").Group("v1.0").Group("policies")
	{
		v10.POST("", middleware.Verify(), middleware.CheckPermission(), control.Create)
		v10.GET("", middleware.Verify(), middleware.CheckPermission(), control.GetByList)
		v10.DELETE("", middleware.Verify(), middleware.CheckPermission(), control.Delete)
	}

	return router
}
