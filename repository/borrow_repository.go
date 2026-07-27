package repository

import (
	"fmt"

	"library-management-backend/constants"
	"library-management-backend/database"
	"library-management-backend/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BorrowRecordWithNames struct {
	model.BorrowRecord
	UserName  string `gorm:"column:user_name"`
	UserUUID  string `gorm:"column:user_uuid"`
	BookTitle string `gorm:"column:book_title"`
	BookUUID  string `gorm:"column:book_uuid"`
}

type BorrowRepository interface {
	StoreBorrowRecordWithtx(tx *gorm.DB, record *model.BorrowRecord) (*model.BorrowRecord, error)
	CountActiveBorrowsByUser(userID uint) (int64, error)
	HasActiveBorrowForBook(userID uint, bookID uint) (bool, error)
	DecrementAvailableCopiesWithtx(tx *gorm.DB, bookID uint) error
	GetBorrowRecordByUUID(uuid string) (*BorrowRecordWithNames, error)
}

type borrowRepository struct {
	db *gorm.DB
}

func NewBorrowRepository() BorrowRepository {
	return &borrowRepository{db: database.LibraryManagementDB}
}

func (r *borrowRepository) getDB() *gorm.DB {
	if r.db != nil {
		return r.db
	}
	return database.LibraryManagementDB
}

func (r *borrowRepository) StoreBorrowRecordWithtx(tx *gorm.DB, record *model.BorrowRecord) (*model.BorrowRecord, error) {
	if tx == nil {
		tx = r.getDB()
	}
	result := tx.Create(record)
	if result.Error != nil {
		logrus.Errorf("Error creating borrow record: %v", result.Error)
		return nil, result.Error
	}
	logrus.Infof("Borrow record created successfully: %+v", record)
	return record, nil
}

func (r *borrowRepository) CountActiveBorrowsByUser(userID uint) (int64, error) {
	var count int64
	err := r.getDB().Model(&model.BorrowRecord{}).
		Where("user_id = ? AND status IN (?, ?) AND deleted_at IS NULL", userID, constants.StatusBorrowed, constants.StatusOverdue).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *borrowRepository) HasActiveBorrowForBook(userID uint, bookID uint) (bool, error) {
	var count int64
	err := r.getDB().Model(&model.BorrowRecord{}).
		Where("user_id = ? AND book_id = ? AND status IN (?, ?) AND deleted_at IS NULL", userID, bookID, constants.StatusBorrowed, constants.StatusOverdue).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *borrowRepository) DecrementAvailableCopiesWithtx(tx *gorm.DB, bookID uint) error {
	if tx == nil {
		tx = r.getDB()
	}
	result := tx.Model(&model.Book{}).
		Where("id = ? AND available_copies > 0 AND deleted_at IS NULL", bookID).
		Update("available_copies", gorm.Expr("available_copies - 1"))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no available copies or book not found for ID %d", bookID)
	}
	return nil
}

func (r *borrowRepository) GetBorrowRecordByUUID(uuid string) (*BorrowRecordWithNames, error) {
	var res BorrowRecordWithNames
	err := r.getDB().Table("borrow_records").
		Select("borrow_records.*, u.name as user_name, u.uuid as user_uuid, b.title as book_title, b.uuid as book_uuid").
		Joins("LEFT JOIN users u ON borrow_records.user_id = u.id").
		Joins("LEFT JOIN books b ON borrow_records.book_id = b.id").
		Where("borrow_records.uuid = ? AND borrow_records.deleted_at IS NULL", uuid).
		Scan(&res).Error
	if err != nil {
		return nil, err
	}
	if res.ID == 0 {
		return nil, nil
	}
	return &res, nil
}
