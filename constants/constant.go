package constants

// Roles
type Role string

const (
	RoleAdmin  Role = "ADMIN"
	RoleMember Role = "MEMBER"
)

// User Statuses
type UserStatus string

const (
	StatusActive   UserStatus = "ACTIVE"
	StatusInactive UserStatus = "INACTIVE"
)

// Borrow Statuses
type BorrowStatus string

const (
	StatusBorrowed BorrowStatus = "BORROWED"
	StatusReturned BorrowStatus = "RETURNED"
	StatusOverdue  BorrowStatus = "OVERDUE"
)
