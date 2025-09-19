package router

import (
	"upload-util/internal/config"
	"upload-util/internal/handler"
	"upload-util/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.UploadConfig) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.Cors())

	uploadHandler, err := handler.NewUploadHandler(cfg)
	if err != nil {
		return nil, err
	}
	api := r.Group("/api/v1")
	{
		upload := api.Group("/upload")
		{
			upload.POST("/file", uploadHandler.Upload)
			upload.POST("files", uploadHandler.UploadMultiple)
			upload.GET("/url", uploadHandler.GetURL)
			upload.DELETE("/file", uploadHandler.Delete)
		}

		system := api.Group("/system")
		{
			system.GET("/health", uploadHandler.HealthCheck)
		}
	}
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/api/v1/system/health")
	})
	return r, nil
}
