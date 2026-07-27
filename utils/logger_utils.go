package utils

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var currentLogDate string // Store the current log date

func InitializeLogger() {
	logDir := "./log"
	currentLogDate = time.Now().Format("2006-01-02")
	logFileName := getLogFileName()

	if err := createLogDirectory(logDir); err != nil {
		logrus.Error("Failed to create log directory:", err)
		return
	}

	file, err := os.OpenFile(path.Join(logDir, logFileName), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		logrus.Error("Failed to open log file:", err)
		return
	}

	logrus.SetOutput(file)
	logrus.SetFormatter(&logrus.JSONFormatter{
		DisableHTMLEscape: true,
		PrettyPrint:       true,
		TimestampFormat:   "2006-01-02 15:04:05",
	})
	gin.DefaultWriter = io.MultiWriter(file)
	logrus.SetReportCaller(true)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Debug("logFileName: ", logFileName)
}

func getLogFileName() string {
	now := time.Now()
	currentDate := now.Format("2006-01-02")
	if currentDate != currentLogDate {
		currentLogDate = currentDate
		createLogFile()
	}
	return currentDate + ".log"
}

func createLogFile() {
	logDir := "./log"
	logFileName := getLogFileName()
	file, err := os.OpenFile(path.Join(logDir, logFileName), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		logrus.Error("Failed to open log file:", err)
		return
	}

	logrus.SetOutput(file)
	logrus.Debug("Created new log file: ", logFileName)
}

func createLogDirectory(logDir string) error {
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
