package main

import (
	"log"
	"net/http"

	"library-management-backend/config"
	"library-management-backend/database"

	"github.com/gin-gonic/gin"
)

func main() {

	gin.SetMode(gin.ReleaseMode)

	cfg := config.LoadConfig()
	db := database.ConnectDB(cfg)
	_ = db

	router := gin.New()
	router.Use(gin.Recovery())
	router.SetTrustedProxies(nil)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Library Management API is running",
		})
	})
	log.Printf("🚀 Server running on port %s", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
