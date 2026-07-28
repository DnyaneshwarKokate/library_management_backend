package dto

import (
	"time"

	"library-management-backend/constants"
)

type BorrowBookRequest struct {
	BookUUID string `json:"book_uuid" validate:"required" binding:"required"`
}

type ReturnBookRequest struct {
	BorrowRecordUUID string `json:"borrow_record_uuid,omitempty"`
}

type BorrowHistoryFilter struct {
	UserID   uint   `json:"user_id,omitempty"`
	Status   string `json:"status,omitempty"`
	FromDate string `json:"from_date,omitempty"`
	ToDate   string `json:"to_date,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

type BorrowRecordResponse struct {
	ID         uint                   `json:"id"`
	UUID       string                 `json:"uuid"`
	UserID     uint                   `json:"user_id"`
	UserUUID   string                 `json:"user_uuid,omitempty"`
	UserName   string                 `json:"user_name,omitempty"`
	BookID     uint                   `json:"book_id"`
	BookUUID   string                 `json:"book_uuid,omitempty"`
	BookTitle  string                 `json:"book_title,omitempty"`
	BorrowDate time.Time              `json:"borrow_date"`
	DueDate    time.Time              `json:"due_date"`
	ReturnedAt *time.Time             `json:"returned_at,omitempty"`
	Status     constants.BorrowStatus `json:"status"`
	CreatedBy  *uint                  `json:"created_by,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}
