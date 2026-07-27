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

type BookWithUserNames struct {
	model.Book
	CreatedByName string `gorm:"column:created_by_name"`
	UpdatedByName string `gorm:"column:updated_by_name"`
}

type BookRepository interface {
	StoreBookWithtx(tx *gorm.DB, book *model.Book) (*model.Book, error)
	GetBookByUUID(uuid string) (*BookWithUserNames, error)
	GetBookList(filter dto.BookFilter) ([]BookWithUserNames, int64, int64, error)
	UpdateBookWithtx(tx *gorm.DB, book *model.Book) (*model.Book, error)
	DeleteBookByUUID(uuid string) error
	ExistsByISBN(isbn string) (bool, error)
	HasActiveBorrowRecords(bookID uint) (bool, error)
}

type bookRepository struct {
	db *gorm.DB
}

func NewBookRepository() BookRepository {
	return &bookRepository{db: database.LibraryManagementDB}
}

func (r *bookRepository) getDB() *gorm.DB {
	if r.db != nil {
		return r.db
	}
	return database.LibraryManagementDB
}

func (r *bookRepository) StoreBookWithtx(tx *gorm.DB, book *model.Book) (*model.Book, error) {
	if tx == nil {
		tx = r.getDB()
	}
	result := tx.Create(book)
	if result.Error != nil {
		logrus.Errorf("Error creating book: %v", result.Error)
		return nil, result.Error
	}
	logrus.Infof("Book created successfully: %+v", book)
	return book, nil
}

func (r *bookRepository) GetBookByUUID(uuid string) (*BookWithUserNames, error) {
	var res BookWithUserNames
	err := r.getDB().Table("books").
		Select("books.*, cu.name as created_by_name, uu.name as updated_by_name").
		Joins("LEFT JOIN users cu ON books.created_by = cu.id").
		Joins("LEFT JOIN users uu ON books.updated_by = uu.id").
		Where("books.uuid = ? AND books.deleted_at IS NULL", uuid).
		Scan(&res).Error
	if err != nil {
		return nil, err
	}
	if res.ID == 0 {
		return nil, nil
	}
	return &res, nil
}

func (r *bookRepository) ExistsByISBN(isbn string) (bool, error) {
	var count int64
	err := r.getDB().Model(&model.Book{}).Where("isbn = ? AND deleted_at IS NULL", isbn).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *bookRepository) GetBookList(filter dto.BookFilter) ([]BookWithUserNames, int64, int64, error) {
	var books []BookWithUserNames
	var totalCount int64
	var filteredCount int64

	// Total count of all non-deleted books
	if err := r.getDB().Table("books").Where("deleted_at IS NULL").Count(&totalCount).Error; err != nil {
		return nil, 0, 0, err
	}

	query := r.getDB().Table("books").Where("books.deleted_at IS NULL")

	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		query = query.Where("books.title LIKE ? OR books.author LIKE ?", searchTerm, searchTerm)
	}

	if filter.Category != "" {
		query = query.Where("books.category = ?", filter.Category)
	}

	if err := query.Count(&filteredCount).Error; err != nil {
		return nil, 0, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	err := query.
		Select("books.*, cu.name as created_by_name, uu.name as updated_by_name").
		Joins("LEFT JOIN users cu ON books.created_by = cu.id").
		Joins("LEFT JOIN users uu ON books.updated_by = uu.id").
		Order("books.id desc").
		Limit(limit).
		Offset(offset).
		Scan(&books).Error
	if err != nil {
		return nil, 0, 0, err
	}

	return books, totalCount, filteredCount, nil
}

func (r *bookRepository) UpdateBookWithtx(tx *gorm.DB, book *model.Book) (*model.Book, error) {
	if tx == nil {
		tx = r.getDB()
	}
	result := tx.Save(book)
	if result.Error != nil {
		logrus.Errorf("Error updating book: %v", result.Error)
		return nil, result.Error
	}
	logrus.Infof("Book updated successfully: %+v", book)
	return book, nil
}

func (r *bookRepository) DeleteBookByUUID(uuid string) error {
	now := time.Now()
	result := r.getDB().Model(&model.Book{}).Where("uuid = ? AND deleted_at IS NULL", uuid).Update("deleted_at", now)
	if result.Error != nil {
		logrus.Errorf("Error deleting book by UUID: %v", result.Error)
		return result.Error
	}
	return nil
}

func (r *bookRepository) HasActiveBorrowRecords(bookID uint) (bool, error) {
	var count int64
	err := r.getDB().Model(&model.BorrowRecord{}).
		Where("book_id = ? AND status IN (?, ?) AND deleted_at IS NULL", bookID, constants.StatusBorrowed, constants.StatusOverdue).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
