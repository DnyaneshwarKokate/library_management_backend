package repository

import (
	"time"

	"library-management-backend/constants"
	"library-management-backend/database"
	"library-management-backend/dto"
	"library-management-backend/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetDashboardStats() (*dto.AdminDashboardResponse, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository() DashboardRepository {
	return &dashboardRepository{db: database.LibraryManagementDB}
}

func (r *dashboardRepository) getDB() *gorm.DB {
	if r.db != nil {
		return r.db
	}
	return database.LibraryManagementDB
}

func (r *dashboardRepository) GetDashboardStats() (*dto.AdminDashboardResponse, error) {
	db := r.getDB()
	var stats dto.AdminDashboardResponse

	if err := db.Model(&model.Book{}).Where("deleted_at IS NULL").Count(&stats.TotalBooks).Error; err != nil {
		logrus.Error("GetDashboardStats TotalBooks DB Error: ", err)
		return nil, err
	}
	if err := db.Model(&model.User{}).Where("deleted_at IS NULL").Count(&stats.TotalUsers).Error; err != nil {
		logrus.Error("GetDashboardStats TotalUsers DB Error: ", err)
		return nil, err
	}
	var totalAvailable *int64
	if err := db.Model(&model.Book{}).Where("deleted_at IS NULL").Select("COALESCE(SUM(available_copies), 0)").Scan(&totalAvailable).Error; err != nil {
		logrus.Error("GetDashboardStats TotalAvailableBooks DB Error: ", err)
		return nil, err
	}
	if totalAvailable != nil {
		stats.TotalAvailableBooks = *totalAvailable
	}
	if err := db.Model(&model.BorrowRecord{}).
		Where("status = ? AND deleted_at IS NULL", constants.StatusBorrowed).
		Count(&stats.ActiveBorrowings).Error; err != nil {
		logrus.Error("GetDashboardStats ActiveBorrowings DB Error: ", err)
		return nil, err
	}
	now := time.Now()
	if err := db.Model(&model.BorrowRecord{}).
		Where("(status = ? OR (status = ? AND due_date < ?)) AND deleted_at IS NULL", constants.StatusOverdue, constants.StatusBorrowed, now).
		Count(&stats.OverdueBooks).Error; err != nil {
		logrus.Error("GetDashboardStats OverdueBooks DB Error: ", err)
		return nil, err
	}
	if err := db.Model(&model.BorrowRecord{}).
		Where("status = ? AND deleted_at IS NULL", constants.StatusReturned).
		Count(&stats.CompletedBorrowings).Error; err != nil {
		logrus.Error("GetDashboardStats CompletedBorrowings DB Error: ", err)
		return nil, err
	}

	return &stats, nil
}
