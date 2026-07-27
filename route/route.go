package route

import (
	"net/http"

	"library-management-backend/app"

	"github.com/gin-gonic/gin"
)

func SetupRouter(ctl *app.App) *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Library Management API is running",
		})
	})

	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/register", ctl.ConsentRequestController.AuthController.Register)
		authGroup.POST("/login", ctl.ConsentRequestController.AuthController.Login)
	}

	return router
}
