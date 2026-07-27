package database

import (
	"library-management-backend/config"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var LibraryManagementDB *gorm.DB

func InitLibraryManagementDB() error {
	dbConfig := config.GetPrimaryMySQLDBConfig()
	dsn := dbConfig.Username + ":" + dbConfig.Password + "@tcp(" + dbConfig.Host + ":" + dbConfig.Port + ")/" + dbConfig.Database + "?charset=utf8mb4&parseTime=True&loc=Local&sql_mode=''"
	
	var err error
	LibraryManagementDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Error("Failed to connect to MySQL database:", err)
		return err
	}

	sqlDB, err := LibraryManagementDB.DB()
	if err != nil {
		logrus.Error("Failed to configure SQL database pool:", err)
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logrus.Info("Successfully connected to MySQL database!")
	return nil
}
