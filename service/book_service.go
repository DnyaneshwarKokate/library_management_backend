package service

import (
	"errors"

	"library-management-backend/dto"
	"library-management-backend/model"
	"library-management-backend/repository"
	"library-management-backend/utils"

	"github.com/sirupsen/logrus"
)

var (
	ErrDuplicateISBN            = errors.New("book with this ISBN already exists")
	ErrBookNotFound             = errors.New("book not found")
	ErrInvalidAvailableCopies   = errors.New("available copies cannot be negative")
	ErrActiveBorrowRecordsExist = errors.New("cannot delete book with active borrow records")
)

type BookService interface {
	CreateBook(req dto.CreateBookRequest, adminID uint) (*dto.BookResponse, error)
	GetBooksList(filter dto.BookFilter) ([]dto.BookResponse, int64, int64, error)
	GetBookByUUID(uuid string) (*dto.BookResponse, error)
	UpdateBook(uuid string, req dto.UpdateBookRequest, adminID uint) (*dto.BookResponse, error)
	DeleteBook(uuid string) error
}

type bookService struct {
	bookRepo repository.BookRepository
}

func NewBookService(bookRepo repository.BookRepository) BookService {
	return &bookService{
		bookRepo: bookRepo,
	}
}

func (s *bookService) CreateBook(req dto.CreateBookRequest, adminID uint) (*dto.BookResponse, error) {
	logrus.Info("CreateBook@Service Started")

	exists, err := s.bookRepo.ExistsByISBN(req.ISBN)
	if err != nil {
		logrus.Errorf("CreateBook@Service ExistsByISBN Error: %v", err)
		return nil, err
	}
	if exists {
		logrus.Warnf("CreateBook@Service Duplicate ISBN: %s", req.ISBN)
		return nil, ErrDuplicateISBN
	}

	var createdBy *uint
	if adminID > 0 {
		createdBy = &adminID
	}

	book := &model.Book{
		UUID:            utils.WithoutHypenGenUUID(),
		Title:           req.Title,
		Author:          req.Author,
		ISBN:            req.ISBN,
		Category:        req.Category,
		TotalCopies:     req.TotalCopies,
		AvailableCopies: req.TotalCopies,
		CreatedBy:       createdBy,
	}

	savedBook, err := s.bookRepo.StoreBookWithtx(nil, book)
	if err != nil {
		logrus.Errorf("CreateBook@Service StoreBook Error: %v", err)
		return nil, err
	}

	bookWithNames, err := s.bookRepo.GetBookByUUID(savedBook.UUID)
	if err == nil && bookWithNames != nil {
		logrus.Infof("CreateBook@Service Completed successfully for Book ID: %d", savedBook.ID)
		return s.toBookResponse(bookWithNames), nil
	}

	logrus.Infof("CreateBook@Service Completed successfully for Book ID: %d", savedBook.ID)
	return s.toBookResponse(&repository.BookWithUserNames{Book: *savedBook}), nil
}

func (s *bookService) GetBooksList(filter dto.BookFilter) ([]dto.BookResponse, int64, int64, error) {
	logrus.Info("GetBooksList@Service Started")

	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	books, totalCount, filteredCount, err := s.bookRepo.GetBookList(filter)
	if err != nil {
		logrus.Errorf("GetBooksList@Service GetBookList Error: %v", err)
		return nil, 0, 0, err
	}

	bookResponses := make([]dto.BookResponse, 0, len(books))
	for _, book := range books {
		bookResponses = append(bookResponses, *s.toBookResponse(&book))
	}

	logrus.Infof("GetBooksList@Service Completed successfully, TotalCount: %d, FilteredCount: %d", totalCount, filteredCount)
	return bookResponses, totalCount, filteredCount, nil
}

func (s *bookService) GetBookByUUID(uuid string) (*dto.BookResponse, error) {
	logrus.Info("GetBookByUUID@Service Started")

	book, err := s.bookRepo.GetBookByUUID(uuid)
	if err != nil {
		logrus.Errorf("GetBookByUUID@Service GetBookByUUID Error: %v", err)
		return nil, err
	}
	if book == nil {
		logrus.Warnf("GetBookByUUID@Service Book Not Found UUID: %s", uuid)
		return nil, ErrBookNotFound
	}

	logrus.Infof("GetBookByUUID@Service Completed successfully for UUID: %s", uuid)
	return s.toBookResponse(book), nil
}

func (s *bookService) UpdateBook(uuid string, req dto.UpdateBookRequest, adminID uint) (*dto.BookResponse, error) {
	logrus.Info("UpdateBook@Service Started")

	existingBookWithNames, err := s.bookRepo.GetBookByUUID(uuid)
	if err != nil {
		logrus.Errorf("UpdateBook@Service GetBookByUUID Error: %v", err)
		return nil, err
	}
	if existingBookWithNames == nil {
		logrus.Warnf("UpdateBook@Service Book Not Found UUID: %s", uuid)
		return nil, ErrBookNotFound
	}

	existingBook := &existingBookWithNames.Book
	copyDiff := req.TotalCopies - existingBook.TotalCopies
	newAvailableCopies := existingBook.AvailableCopies + copyDiff
	if newAvailableCopies < 0 {
		logrus.Warnf("UpdateBook@Service Invalid Available Copies for UUID: %s", uuid)
		return nil, ErrInvalidAvailableCopies
	}

	var updatedBy *uint
	if adminID > 0 {
		updatedBy = &adminID
	}

	existingBook.Title = req.Title
	existingBook.Author = req.Author
	existingBook.Category = req.Category
	existingBook.TotalCopies = req.TotalCopies
	existingBook.AvailableCopies = newAvailableCopies
	existingBook.UpdatedBy = updatedBy

	updatedBook, err := s.bookRepo.UpdateBookWithtx(nil, existingBook)
	if err != nil {
		logrus.Errorf("UpdateBook@Service UpdateBook Error: %v", err)
		return nil, err
	}

	bookWithNames, err := s.bookRepo.GetBookByUUID(updatedBook.UUID)
	if err == nil && bookWithNames != nil {
		logrus.Infof("UpdateBook@Service Completed successfully for Book ID: %d", updatedBook.ID)
		return s.toBookResponse(bookWithNames), nil
	}

	logrus.Infof("UpdateBook@Service Completed successfully for Book ID: %d", updatedBook.ID)
	return s.toBookResponse(&repository.BookWithUserNames{Book: *updatedBook}), nil
}

func (s *bookService) DeleteBook(uuid string) error {
	logrus.Info("DeleteBook@Service Started")

	existingBook, err := s.bookRepo.GetBookByUUID(uuid)
	if err != nil {
		logrus.Errorf("DeleteBook@Service GetBookByUUID Error: %v", err)
		return err
	}
	if existingBook == nil {
		logrus.Warnf("DeleteBook@Service Book Not Found UUID: %s", uuid)
		return ErrBookNotFound
	}

	hasActive, err := s.bookRepo.HasActiveBorrowRecords(existingBook.ID)
	if err != nil {
		logrus.Errorf("DeleteBook@Service HasActiveBorrowRecords Error: %v", err)
		return err
	}
	if hasActive {
		logrus.Warnf("DeleteBook@Service Active Borrow Records Exist for Book ID: %d", existingBook.ID)
		return ErrActiveBorrowRecordsExist
	}

	err = s.bookRepo.DeleteBookByUUID(uuid)
	if err != nil {
		logrus.Errorf("DeleteBook@Service DeleteBookByUUID Error: %v", err)
		return err
	}

	logrus.Infof("DeleteBook@Service Completed successfully for UUID: %s", uuid)
	return nil
}

func (s *bookService) toBookResponse(item *repository.BookWithUserNames) *dto.BookResponse {
	return &dto.BookResponse{
		ID:              item.ID,
		UUID:            item.UUID,
		Title:           item.Title,
		Author:          item.Author,
		ISBN:            item.ISBN,
		Category:        item.Category,
		TotalCopies:     item.TotalCopies,
		AvailableCopies: item.AvailableCopies,
		CreatedBy:       item.CreatedBy,
		CreatedByName:   item.CreatedByName,
		UpdatedBy:       item.UpdatedBy,
		UpdatedByName:   item.UpdatedByName,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}
