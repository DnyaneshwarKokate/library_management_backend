package route

import (
	"net/http"
	"os"

	"library-management-backend/app"
	"library-management-backend/constants"
	"library-management-backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(ctl *app.App) *gin.Engine {
	router := gin.Default()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super_secret_jwt_key_library_app"
	}

	router.POST("/health", func(c *gin.Context) {
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

	bookGroup := router.Group("/api/books")
	{
		bookGroup.POST("/list", ctl.ConsentRequestController.BookController.GetBooks)
		bookGroup.POST("/details", ctl.ConsentRequestController.BookController.GetBookByUUID)
		bookGroup.POST("/details/:uuid", ctl.ConsentRequestController.BookController.GetBookByUUID)

		adminBookGroup := bookGroup.Group("", middleware.AuthMiddleware(jwtSecret), middleware.RequireRole(constants.RoleAdmin))
		{
			adminBookGroup.POST("/create", ctl.ConsentRequestController.BookController.CreateBook)
			adminBookGroup.POST("/update", ctl.ConsentRequestController.BookController.UpdateBook)
			adminBookGroup.POST("/update/:uuid", ctl.ConsentRequestController.BookController.UpdateBook)
			adminBookGroup.POST("/delete", ctl.ConsentRequestController.BookController.DeleteBook)
			adminBookGroup.POST("/delete/:uuid", ctl.ConsentRequestController.BookController.DeleteBook)
		}
	}

	return router
}
