package main

import (
	"flag"
	"log"
	"os"

	"library-management-backend/app"
	"library-management-backend/database"
	"library-management-backend/route"
	"library-management-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func init() {
	envFile := flag.String("env", ".env", "specify the env file name")
	flag.Parse()
	if err := godotenv.Load(*envFile); err != nil {
		logrus.Printf("Notice: Could not load %s file: %v", *envFile, err)
	}
	utils.InitializeLogger()
	logrus.Infof("Environment variables loaded successfully.")
}

func main() {
	if err := database.InitLibraryManagementDB(); err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)

	application := app.InitApp()
	router := route.SetupRouter(application)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server running on port %s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
