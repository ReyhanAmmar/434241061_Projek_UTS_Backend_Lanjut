package model

import "time"

type Cursor struct {
	Date time.Time
	ID   int
}

type CursorMeta struct {
	Limit      int     `json:"limit"`
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor"`
}

type CursorResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Data    any        `json:"data"`
	Meta    CursorMeta `json:"meta"`
}

type Period struct {
	From *time.Time `json:"from"`
	To   *time.Time `json:"to"`
}

type CategoryTotal struct {
	CategoryID int    `json:"category_id"`
	Category   string `json:"category"`
	Type       string `json:"type"`
	Total      int64  `json:"total"`
	Count      int    `json:"count"`
}

type DayTotal struct {
	Date    string `json:"date"`
	Income  int64  `json:"income"`
	Expense int64  `json:"expense"`
	Count   int    `json:"count"`
}

type TransactionSummary struct {
	Period           Period          `json:"period"`
	TotalIncome      int64           `json:"total_income"`
	TotalExpense     int64           `json:"total_expense"`
	Balance          int64           `json:"balance"`
	TransactionCount int             `json:"transaction_count"`
	ByCategory       []CategoryTotal `json:"by_category"`
	ByDay            []DayTotal      `json:"by_day"`
}