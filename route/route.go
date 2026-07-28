package route

import (
	"os"

	"library-management-backend/app"
	"library-management-backend/constants"
	"library-management-backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(ctl *app.ConsentRequestController) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT", "PATCH"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  false,
		AllowOriginFunc:  func(origin string) bool { return true },
		MaxAge:           86400,
	}))

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super_secret_jwt_key_library_app"
	}

	router.POST("/health-check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "Library Management API is running",
		})
	})

	// User Routes
	usersGroup := router.Group("/user")
	{
		usersGroup.POST("/register", ctl.AuthController.Register)
		usersGroup.POST("/login", ctl.AuthController.Login)
	}

	// Books Routes
	booksGroup := router.Group("/books")
	{
		booksGroup.POST("/list", ctl.BookController.GetBooks)
		booksGroup.POST("/details", ctl.BookController.GetBookByUUID)
		booksGroup.POST("/create", middleware.AuthMiddleware(jwtSecret), middleware.RequireRole(constants.RoleAdmin), ctl.BookController.CreateBook)
		booksGroup.POST("/update", middleware.AuthMiddleware(jwtSecret), middleware.RequireRole(constants.RoleAdmin), ctl.BookController.UpdateBook)
		booksGroup.POST("/delete", middleware.AuthMiddleware(jwtSecret), middleware.RequireRole(constants.RoleAdmin), ctl.BookController.DeleteBook)
	}

	// Borrow Routes
	borrowGroup := router.Group("/borrow", middleware.AuthMiddleware(jwtSecret))
	{
		borrowGroup.POST("/create", ctl.BorrowController.BorrowBook)
		borrowGroup.POST("/return", ctl.BorrowController.ReturnBook)
		borrowGroup.POST("/return/:id", ctl.BorrowController.ReturnBook)
	}

	// Borrow Records Routes
	borrowRecordsGroup := router.Group("/borrow-records", middleware.AuthMiddleware(jwtSecret))
	{
		borrowRecordsGroup.POST("/:id/return", ctl.BorrowController.ReturnBook)
	}

	return router
}
