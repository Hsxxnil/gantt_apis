package policy

import (
	"gorm.io/gorm"
	present "hta/internal/presenter/policy"
	"hta/internal/router/middleware"
	"hta/internal/router/middleware/auth"

	"github.com/gin-gonic/gin"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init()
	v10 := router.Group("hta-gantt").Group("v1.0").Group("policies")
	{
		v10.POST("", middleware.Verify(), auth.CheckPermission(), control.Create)
		v10.GET("", middleware.Verify(), auth.CheckPermission(), control.GetByList)
		v10.DELETE("", middleware.Verify(), auth.CheckPermission(), control.Delete)
	}

	return router
}
