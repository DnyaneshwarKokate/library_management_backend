package dto

type AdminDashboardResponse struct {
	TotalBooks          int64 `json:"total_books"`
	TotalUsers          int64 `json:"total_users"`
	TotalAvailableBooks int64 `json:"total_available_books"`
	ActiveBorrowings    int64 `json:"active_borrowings"`
	OverdueBooks        int64 `json:"overdue_books"`
	CompletedBorrowings int64 `json:"completed_borrowings"`
}
