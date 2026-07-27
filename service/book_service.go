package service

import (
	"errors"

	"library-management-backend/dto"
	"library-management-backend/model"
	"library-management-backend/repository"
	"library-management-backend/utils"
)

var (
	ErrDuplicateISBN            = errors.New("book with this ISBN already exists")
	ErrBookNotFound             = errors.New("book not found")
	ErrInvalidAvailableCopies   = errors.New("available copies cannot be negative")
	ErrActiveBorrowRecordsExist = errors.New("cannot delete book with active borrow records")
)

type BookService interface {
	CreateBook(req dto.CreateBookRequest, adminID uint) (*dto.BookResponse, error)
	GetBooks(filter dto.BookFilter) (*dto.BookListResponse, error)
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
	exists, err := s.bookRepo.ExistsByISBN(req.ISBN)
	if err != nil {
		return nil, err
	}
	if exists {
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
		return nil, err
	}

	// Refetch with LEFT JOIN to populate user names in response
	bookWithNames, err := s.bookRepo.GetBookByUUID(savedBook.UUID)
	if err == nil && bookWithNames != nil {
		return s.toBookResponse(bookWithNames), nil
	}

	return s.toBookResponse(&repository.BookWithUserNames{Book: *savedBook}), nil
}

func (s *bookService) GetBooks(filter dto.BookFilter) (*dto.BookListResponse, error) {
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	books, totalCount, filteredCount, err := s.bookRepo.GetBookList(filter)
	if err != nil {
		return nil, err
	}

	bookResponses := make([]dto.BookResponse, 0, len(books))
	for _, book := range books {
		bookResponses = append(bookResponses, *s.toBookResponse(&book))
	}

	return &dto.BookListResponse{
		TotalCount:    totalCount,
		FilteredCount: filteredCount,
		Data:          bookResponses,
	}, nil
}

func (s *bookService) GetBookByUUID(uuid string) (*dto.BookResponse, error) {
	book, err := s.bookRepo.GetBookByUUID(uuid)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, ErrBookNotFound
	}
	return s.toBookResponse(book), nil
}

func (s *bookService) UpdateBook(uuid string, req dto.UpdateBookRequest, adminID uint) (*dto.BookResponse, error) {
	existingBookWithNames, err := s.bookRepo.GetBookByUUID(uuid)
	if err != nil {
		return nil, err
	}
	if existingBookWithNames == nil {
		return nil, ErrBookNotFound
	}

	existingBook := &existingBookWithNames.Book
	copyDiff := req.TotalCopies - existingBook.TotalCopies
	newAvailableCopies := existingBook.AvailableCopies + copyDiff
	if newAvailableCopies < 0 {
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
		return nil, err
	}

	// Refetch with LEFT JOIN to populate updated user names
	bookWithNames, err := s.bookRepo.GetBookByUUID(updatedBook.UUID)
	if err == nil && bookWithNames != nil {
		return s.toBookResponse(bookWithNames), nil
	}

	return s.toBookResponse(&repository.BookWithUserNames{Book: *updatedBook}), nil
}

func (s *bookService) DeleteBook(uuid string) error {
	existingBook, err := s.bookRepo.GetBookByUUID(uuid)
	if err != nil {
		return err
	}
	if existingBook == nil {
		return ErrBookNotFound
	}

	hasActive, err := s.bookRepo.HasActiveBorrowRecords(existingBook.ID)
	if err != nil {
		return err
	}
	if hasActive {
		return ErrActiveBorrowRecordsExist
	}

	return s.bookRepo.DeleteBookByUUID(uuid)
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
