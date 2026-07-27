package model

import (
	"time"
	"library-management-backend/constants"
)

type BorrowRecord struct {
	ID         uint                   `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID       string                 `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`
	UserID     uint                   `gorm:"not null;index" json:"user_id"`
	User       *User                  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"user,omitempty"`
	BookID     uint                   `gorm:"not null;index" json:"book_id"`
	Book       *Book                  `gorm:"foreignKey:BookID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"book,omitempty"`
	BorrowDate time.Time              `gorm:"not null" json:"borrow_date"`
	DueDate    time.Time              `gorm:"not null;index" json:"due_date"`
	ReturnedAt *time.Time             `json:"returned_at,omitempty"`
	Status     constants.BorrowStatus `gorm:"type:enum('BORROWED','RETURNED','OVERDUE');default:'BORROWED';not null;index" json:"status"`
	CreatedBy  *int                   `gorm:"column:created_by" json:"created_by"`
	UpdatedBy  *int                   `gorm:"column:updated_by" json:"updated_by"`
	CreatedAt  time.Time              `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time              `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt  *time.Time             `gorm:"column:deleted_at;type:timestamp;default:NULL" json:"deleted_at"`
}

func (BorrowRecord) TableName() string {
	return "borrow_records"
}
