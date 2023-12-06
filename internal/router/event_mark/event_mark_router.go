package event_mark

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	present "hta/internal/presenter/event_mark"
	"hta/internal/router/middleware"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("hta-gantt").Group("v1.0").Group("event-marks")
	{
		v10.POST("", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Create)
		v10.DELETE(":id", middleware.Verify(), middleware.CheckPermission(), control.Delete)
		v10.PATCH(":id", middleware.Verify(), middleware.CheckPermission(), control.Update)
	}

	return router
}
