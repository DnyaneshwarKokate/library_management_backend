package service

import (
	"errors"
	"time"

	"library-management-backend/constants"
	"library-management-backend/database"
	"library-management-backend/dto"
	"library-management-backend/model"
	"library-management-backend/repository"
	"library-management-backend/utils"
)

var (
	ErrBookOutOfStock        = errors.New("no copies available for this book")
	ErrBookAlreadyBorrowed   = errors.New("you have already borrowed this book")
	ErrBorrowLimitExceeded   = errors.New("maximum active borrow limit of 3 books reached")
	ErrUserNotFound          = errors.New("user not found")
)

type BorrowService interface {
	BorrowBook(req dto.BorrowBookRequest, userID uint) (*dto.BorrowRecordResponse, error)
}

type borrowService struct {
	borrowRepo repository.BorrowRepository
	bookRepo   repository.BookRepository
	userRepo   repository.UserRepository
}

func NewBorrowService(borrowRepo repository.BorrowRepository, bookRepo repository.BookRepository, userRepo repository.UserRepository) BorrowService {
	return &borrowService{
		borrowRepo: borrowRepo,
		bookRepo:   bookRepo,
		userRepo:   userRepo,
	}
}

func (s *borrowService) BorrowBook(req dto.BorrowBookRequest, userID uint) (*dto.BorrowRecordResponse, error) {
	// 1. Verify User
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if user.Status == constants.StatusInactive {
		return nil, ErrUserInactive
	}

	// 2. Verify Book
	bookWithNames, err := s.bookRepo.GetBookByUUID(req.BookUUID)
	if err != nil {
		return nil, err
	}
	if bookWithNames == nil {
		return nil, ErrBookNotFound
	}
	if bookWithNames.AvailableCopies <= 0 {
		return nil, ErrBookOutOfStock
	}

	// 3. Check duplicate borrow for same book
	alreadyBorrowed, err := s.borrowRepo.HasActiveBorrowForBook(userID, bookWithNames.ID)
	if err != nil {
		return nil, err
	}
	if alreadyBorrowed {
		return nil, ErrBookAlreadyBorrowed
	}

	// 4. Check active borrow limit (Max 3 active books)
	activeCount, err := s.borrowRepo.CountActiveBorrowsByUser(userID)
	if err != nil {
		return nil, err
	}
	if activeCount >= 3 {
		return nil, ErrBorrowLimitExceeded
	}

	// 5. Execute ACID Transaction
	tx := database.LibraryManagementDB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.borrowRepo.DecrementAvailableCopiesWithtx(tx, bookWithNames.ID); err != nil {
		tx.Rollback()
		return nil, err
	}

	now := time.Now()
	dueDate := now.AddDate(0, 0, 14) // 14 Days from borrowing

	var createdBy *uint
	if userID > 0 {
		createdBy = &userID
	}

	record := &model.BorrowRecord{
		UUID:       utils.WithoutHypenGenUUID(),
		UserID:     userID,
		BookID:     bookWithNames.ID,
		BorrowDate: now,
		DueDate:    dueDate,
		Status:     constants.StatusBorrowed,
		CreatedBy:  createdBy,
	}

	savedRecord, err := s.borrowRepo.StoreBorrowRecordWithtx(tx, record)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// 6. Refetch with joined names for response
	refetchedRecord, err := s.borrowRepo.GetBorrowRecordByUUID(savedRecord.UUID)
	if err == nil && refetchedRecord != nil {
		return s.toBorrowResponse(refetchedRecord), nil
	}

	return s.toBorrowResponse(&repository.BorrowRecordWithNames{BorrowRecord: *savedRecord}), nil
}

func (s *borrowService) toBorrowResponse(item *repository.BorrowRecordWithNames) *dto.BorrowRecordResponse {
	return &dto.BorrowRecordResponse{
		ID:         item.ID,
		UUID:       item.UUID,
		UserID:     item.UserID,
		UserUUID:   item.UserUUID,
		UserName:   item.UserName,
		BookID:     item.BookID,
		BookUUID:   item.BookUUID,
		BookTitle:  item.BookTitle,
		BorrowDate: item.BorrowDate,
		DueDate:    item.DueDate,
		ReturnedAt: item.ReturnedAt,
		Status:     item.Status,
		CreatedBy:  item.CreatedBy,
		CreatedAt:  item.CreatedAt,
	}
}
