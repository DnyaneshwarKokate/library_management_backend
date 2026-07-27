package constants

// Roles
type Role string

const (
	RoleAdmin  Role = "ADMIN"
	RoleMember Role = "MEMBER"
)

// Borrow Statuses
type BorrowStatus string

const (
	StatusBorrowed BorrowStatus = "BORROWED"
	StatusReturned BorrowStatus = "RETURNED"
	StatusOverdue  BorrowStatus = "OVERDUE"
)
