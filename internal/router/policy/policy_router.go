package policy

import (
	present "hta/internal/presenter/policy"
	"hta/internal/router/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init()
	v10 := router.Group("hta-gantt").Group("v1.0").Group("policies")
	{
		v10.POST("", middleware.Verify(), middleware.CheckPermission(), control.Create)
		v10.GET("", middleware.Verify(), middleware.CheckPermission(), control.GetByList)
		v10.DELETE("", middleware.Verify(), middleware.CheckPermission(), control.Delete)
	}

	return router
}
