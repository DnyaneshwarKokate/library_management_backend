package dto

import "time"

type CreateBookRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255" binding:"required,min=1,max=255"`
	Author      string `json:"author" validate:"required,min=1,max=100" binding:"required,min=1,max=100"`
	ISBN        string `json:"isbn" validate:"required,min=1,max=20" binding:"required,min=1,max=20"`
	Category    string `json:"category" validate:"required,min=1,max=50" binding:"required,min=1,max=50"`
	TotalCopies int    `json:"total_copies" validate:"required,min=1" binding:"required,min=1"`
}

type UpdateBookRequest struct {
	UUID        string `json:"uuid"`
	Title       string `json:"title" validate:"required,min=1,max=255" binding:"required,min=1,max=255"`
	Author      string `json:"author" validate:"required,min=1,max=100" binding:"required,min=1,max=100"`
	Category    string `json:"category" validate:"required,min=1,max=50" binding:"required,min=1,max=50"`
	TotalCopies int    `json:"total_copies" validate:"required,min=1" binding:"required,min=1"`
}

type BookFilter struct {
	Limit    int    `form:"limit" json:"limit"`
	Offset   int    `form:"offset" json:"offset"`
	Search   string `form:"search" json:"search"`
	Category string `form:"category" json:"category"`
}

type BookResponse struct {
	ID              uint      `json:"id"`
	UUID            string    `json:"uuid"`
	Title           string    `json:"title"`
	Author          string    `json:"author"`
	ISBN            string    `json:"isbn"`
	Category        string    `json:"category"`
	TotalCopies     int       `json:"total_copies"`
	AvailableCopies int       `json:"available_copies"`
	CreatedBy       *uint     `json:"created_by,omitempty"`
	CreatedByName   string    `json:"created_by_name,omitempty"`
	UpdatedBy       *uint     `json:"updated_by,omitempty"`
	UpdatedByName   string    `json:"updated_by_name,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type BookListResponse struct {
	TotalCount    int64          `json:"total_count"`
	FilteredCount int64          `json:"filtered_count"`
	Data          []BookResponse `json:"data"`
}
