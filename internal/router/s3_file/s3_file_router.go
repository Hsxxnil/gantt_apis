package s3_file

import (
	present "hta/internal/presenter/s3_file"
	"hta/internal/router/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouter(router *gin.Engine, db *gorm.DB) *gin.Engine {
	control := present.Init(db)
	v10 := router.Group("hta-gantt").Group("v1.0").Group("files")
	{
		v10.POST("", middleware.Verify(), middleware.CheckPermission(), middleware.Transaction(db), control.Create)
		v10.DELETE(":id", middleware.Verify(), middleware.CheckPermission(), control.Delete)
	}

	return router
}
