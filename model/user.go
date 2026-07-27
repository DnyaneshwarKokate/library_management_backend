package model

import (
	"time"

	"library-management-backend/constants"
)

type User struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID          string         `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`
	Name          string         `gorm:"type:varchar(100);not null" json:"name"`
	Email         string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	PasswordHash  string         `gorm:"type:varchar(255);not null" json:"-"`
	Role          constants.Role `gorm:"type:enum('ADMIN','MEMBER');default:'MEMBER';not null" json:"role"`
	BorrowRecords []BorrowRecord `gorm:"foreignKey:UserID" json:"borrow_records,omitempty"`
	CreatedBy     *int           `gorm:"column:created_by" json:"created_by"`
	UpdatedBy     *int           `gorm:"column:updated_by" json:"updated_by"`
	CreatedAt     time.Time      `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt     *time.Time     `gorm:"column:deleted_at;type:timestamp;default:NULL" json:"deleted_at"`
}

func (User) TableName() string {
	return "users"
}
