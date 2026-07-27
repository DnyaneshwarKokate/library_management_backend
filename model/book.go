package model

import (
	"time"
)

type Book struct {
	ID              uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID            string         `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`
	Title           string         `gorm:"type:varchar(255);index;not null" json:"title"`
	Author          string         `gorm:"type:varchar(100);index;not null" json:"author"`
	ISBN            string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"isbn"`
	Category        string         `gorm:"type:varchar(50);index;not null" json:"category"`
	TotalCopies     int            `gorm:"not null" json:"total_copies"`
	AvailableCopies int            `gorm:"not null" json:"available_copies"`
	BorrowRecords   []BorrowRecord `gorm:"foreignKey:BookID" json:"borrow_records,omitempty"`
	CreatedBy       *uint          `gorm:"column:created_by" json:"created_by"`
	UpdatedBy       *uint          `gorm:"column:updated_by" json:"updated_by"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt       *time.Time     `gorm:"column:deleted_at;type:timestamp;default:NULL" json:"deleted_at"`
}

func (Book) TableName() string {
	return "books"
}
