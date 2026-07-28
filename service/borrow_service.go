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
	"library-management-backend/workers"

	"github.com/sirupsen/logrus"
)

var (
	ErrBookOutOfStock       = errors.New("no copies available for this book")
	ErrBookAlreadyBorrowed  = errors.New("you have already borrowed this book")
	ErrBorrowLimitExceeded  = errors.New("maximum active borrow limit of 3 books reached")
	ErrUserNotFound         = errors.New("user not found")
	ErrBorrowRecordNotFound = errors.New("borrow record not found")
	ErrAlreadyReturned      = errors.New("book is already returned")
	ErrUnauthorizedReturn   = errors.New("you are not authorized to return this borrowing record")
)

type BorrowService interface {
	BorrowBook(req dto.BorrowBookRequest, userID uint) (*dto.BorrowRecordResponse, error)
	ReturnBook(recordUUID string, userID uint) (*dto.BorrowRecordResponse, error)
	GetMyBorrowings(filter dto.BorrowHistoryFilter) ([]dto.BorrowRecordResponse, int64, int64, error)
	ProcessOverdue() (*dto.ProcessOverdueResponse, error)
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
	logrus.Info("BorrowBook@Service Started")

	// 1. Verify User
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		logrus.Errorf("BorrowBook@Service FindByID Error: %v", err)
		return nil, err
	}
	if user == nil {
		logrus.Warnf("BorrowBook@Service User Not Found ID: %d", userID)
		return nil, ErrUserNotFound
	}
	if user.Status == constants.StatusInactive {
		logrus.Warnf("BorrowBook@Service Inactive User ID: %d", userID)
		return nil, ErrUserInactive
	}

	// 2. Verify Book
	bookWithNames, err := s.bookRepo.GetBookByUUID(req.BookUUID)
	if err != nil {
		logrus.Errorf("BorrowBook@Service GetBookByUUID Error: %v", err)
		return nil, err
	}
	if bookWithNames == nil {
		logrus.Warnf("BorrowBook@Service Book Not Found UUID: %s", req.BookUUID)
		return nil, ErrBookNotFound
	}
	if bookWithNames.AvailableCopies <= 0 {
		logrus.Warnf("BorrowBook@Service Book Out of Stock Book ID: %d", bookWithNames.ID)
		return nil, ErrBookOutOfStock
	}

	// 3. Check duplicate borrow for same book
	alreadyBorrowed, err := s.borrowRepo.HasActiveBorrowForBook(userID, bookWithNames.ID)
	if err != nil {
		logrus.Errorf("BorrowBook@Service HasActiveBorrowForBook Error: %v", err)
		return nil, err
	}
	if alreadyBorrowed {
		logrus.Warnf("BorrowBook@Service Book Already Borrowed User ID: %d, Book ID: %d", userID, bookWithNames.ID)
		return nil, ErrBookAlreadyBorrowed
	}

	// 4. Check active borrow limit (Max 3 active books)
	activeCount, err := s.borrowRepo.CountActiveBorrowsByUser(userID)
	if err != nil {
		logrus.Errorf("BorrowBook@Service CountActiveBorrowsByUser Error: %v", err)
		return nil, err
	}
	if activeCount >= 3 {
		logrus.Warnf("BorrowBook@Service Borrow Limit Exceeded User ID: %d, Count: %d", userID, activeCount)
		return nil, ErrBorrowLimitExceeded
	}

	// 5. Execute ACID Transaction with Pessimistic Locking (FOR UPDATE)
	tx := database.LibraryManagementDB.Begin()
	if tx.Error != nil {
		logrus.Errorf("BorrowBook@Service transaction begin error: %v", tx.Error)
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logrus.Errorf("BorrowBook@Service panic recovered, transaction rolled back: %v", r)
		}
	}()

	// Lock book row FOR UPDATE to prevent race conditions during concurrent borrowing
	lockedBook, err := s.borrowRepo.GetBookForUpdateWithtx(tx, bookWithNames.ID)
	if err != nil {
		tx.Rollback()
		logrus.Errorf("BorrowBook@Service GetBookForUpdate Error: %v", err)
		return nil, err
	}
	if lockedBook == nil {
		tx.Rollback()
		return nil, ErrBookNotFound
	}
	if lockedBook.AvailableCopies <= 0 {
		tx.Rollback()
		logrus.Warnf("BorrowBook@Service Book Out of Stock inside Lock Book ID: %d", bookWithNames.ID)
		return nil, ErrBookOutOfStock
	}

	if err := s.borrowRepo.DecrementAvailableCopiesWithtx(tx, bookWithNames.ID); err != nil {
		tx.Rollback()
		logrus.Errorf("BorrowBook@Service DecrementAvailableCopies Error: %v", err)
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
		logrus.Errorf("BorrowBook@Service StoreBorrowRecord Error: %v", err)
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		logrus.Errorf("BorrowBook@Service commit failed: %v", err)
		return nil, err
	}

	// 6. Refetch with joined names for response
	refetchedRecord, err := s.borrowRepo.GetBorrowRecordByUUID(savedRecord.UUID)
	if err == nil && refetchedRecord != nil {
		logrus.Infof("BorrowBook@Service Completed successfully for Borrow ID: %d", savedRecord.ID)
		return s.toBorrowResponse(refetchedRecord), nil
	}

	logrus.Infof("BorrowBook@Service Completed successfully for Borrow ID: %d", savedRecord.ID)
	return s.toBorrowResponse(&repository.BorrowRecordWithNames{BorrowRecord: *savedRecord}), nil
}

func (s *borrowService) ReturnBook(recordUUID string, userID uint) (*dto.BorrowRecordResponse, error) {
	logrus.Info("ReturnBook@Service Started")

	// 1. Fetch borrow record
	recordWithNames, err := s.borrowRepo.GetBorrowRecordByUUID(recordUUID)
	if err != nil {
		logrus.Errorf("ReturnBook@Service GetBorrowRecordByUUID Error: %v", err)
		return nil, err
	}
	if recordWithNames == nil {
		logrus.Warnf("ReturnBook@Service Borrow Record Not Found UUID: %s", recordUUID)
		return nil, ErrBorrowRecordNotFound
	}

	// 2. Verify ownership
	if recordWithNames.UserID != userID {
		logrus.Warnf("ReturnBook@Service Unauthorized Return Attempt User ID: %d, Record Owner: %d", userID, recordWithNames.UserID)
		return nil, ErrUnauthorizedReturn
	}

	// 3. Verify status
	if recordWithNames.Status == constants.StatusReturned {
		logrus.Warnf("ReturnBook@Service Already Returned Record UUID: %s", recordUUID)
		return nil, ErrAlreadyReturned
	}

	// 4. Transaction to update status and increment available copies
	tx := database.LibraryManagementDB.Begin()
	if tx.Error != nil {
		logrus.Errorf("ReturnBook@Service transaction begin error: %v", tx.Error)
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logrus.Errorf("ReturnBook@Service panic recovered, transaction rolled back: %v", r)
		}
	}()

	now := time.Now()
	if err := s.borrowRepo.UpdateBorrowRecordStatusWithtx(tx, recordWithNames.ID, constants.StatusReturned, now); err != nil {
		tx.Rollback()
		logrus.Errorf("ReturnBook@Service UpdateBorrowRecordStatus Error: %v", err)
		return nil, err
	}

	if err := s.borrowRepo.IncrementAvailableCopiesWithtx(tx, recordWithNames.BookID); err != nil {
		tx.Rollback()
		logrus.Errorf("ReturnBook@Service IncrementAvailableCopies Error: %v", err)
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		logrus.Errorf("ReturnBook@Service commit failed: %v", err)
		return nil, err
	}

	// 5. Refetch record
	refetchedRecord, err := s.borrowRepo.GetBorrowRecordByUUID(recordUUID)
	if err == nil && refetchedRecord != nil {
		logrus.Infof("ReturnBook@Service Completed successfully for Record ID: %d", recordWithNames.ID)
		return s.toBorrowResponse(refetchedRecord), nil
	}

	recordWithNames.Status = constants.StatusReturned
	recordWithNames.ReturnedAt = &now
	logrus.Infof("ReturnBook@Service Completed successfully for Record ID: %d", recordWithNames.ID)
	return s.toBorrowResponse(recordWithNames), nil
}

func (s *borrowService) GetMyBorrowings(filter dto.BorrowHistoryFilter) ([]dto.BorrowRecordResponse, int64, int64, error) {
	logrus.Info("GetMyBorrowings@Service Started")

	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	records, totalCount, filteredCount, err := s.borrowRepo.GetBorrowHistory(filter)
	if err != nil {
		logrus.Errorf("GetMyBorrowings@Service GetBorrowHistory Error: %v", err)
		return nil, 0, 0, err
	}

	responses := make([]dto.BorrowRecordResponse, 0, len(records))
	for _, rec := range records {
		responses = append(responses, *s.toBorrowResponse(&rec))
	}

	logrus.Infof("GetMyBorrowings@Service Completed successfully, TotalCount: %d, FilteredCount: %d", totalCount, filteredCount)
	return responses, totalCount, filteredCount, nil
}

func (s *borrowService) ProcessOverdue() (*dto.ProcessOverdueResponse, error) {
	logrus.Info("ProcessOverdue@Service Started")

	records, err := s.borrowRepo.GetOverdueRecords()
	if err != nil {
		logrus.Errorf("ProcessOverdue@Service GetOverdueRecords Error: %v", err)
		return nil, err
	}

	res := workers.ProcessOverdueWorkerPool(records, s.borrowRepo.MarkRecordAsOverdue)

	logrus.Infof("ProcessOverdue@Service Completed successfully: %+v", res)
	return res, nil
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
