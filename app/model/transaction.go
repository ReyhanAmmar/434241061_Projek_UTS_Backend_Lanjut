package model

import "time"

const (
	TransactionIncome  = "income"
	TransactionExpense = "expense"
)

type Transaction struct {
	ID              int       `json:"id"`
	Type            string    `json:"type"`
	Amount          int64     `json:"amount"`
	Note            string    `json:"note"`
	TransactionDate time.Time `json:"transaction_date"`
	CategoryID      int       `json:"category_id"`
	OwnerID         int       `json:"owner_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateTransactionRequest struct {
	Type   string `json:"type"   validate:"required,oneof=income expense"`
	Amount int64  `json:"amount" validate:"required,gt=0,lte=1000000000000"`
	Note   string `json:"note"   validate:"max=255"`
	TransactionDate *time.Time `json:"transaction_date" validate:"omitnil,notfuture"`
	CategoryID      int        `json:"category_id"      validate:"required,gt=0"`
}

type ReplaceTransactionRequest struct {
	Type            string    `json:"type"             validate:"required,oneof=income expense"`
	Amount          int64     `json:"amount"           validate:"required,gt=0,lte=1000000000000"`
	Note            string    `json:"note"             validate:"max=255"`
	TransactionDate time.Time `json:"transaction_date" validate:"required,notfuture"`
	CategoryID      int       `json:"category_id"      validate:"required,gt=0"`
}

type PatchTransactionRequest struct {
	Type            *string    `json:"type"             validate:"omitnil,oneof=income expense"`
	Amount          *int64     `json:"amount"           validate:"omitnil,gt=0,lte=1000000000000"`
	Note            *string    `json:"note"             validate:"omitnil,max=255"`
	TransactionDate *time.Time `json:"transaction_date" validate:"omitnil,notfuture"`
	CategoryID      *int       `json:"category_id"      validate:"omitnil,gt=0"`
}

type TransactionQuery struct {
	OwnerID    *int 
	Type       string
	CategoryID *int
	From       *time.Time 
	To         *time.Time 
	Search     string
	Sort       string
	Order      string
	Page       int
	Limit      int
}

func (q TransactionQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}