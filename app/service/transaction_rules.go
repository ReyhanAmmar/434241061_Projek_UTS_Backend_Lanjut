package service

import "api-buku-kas/app/model"

func IsEmptyTransactionPatch(req model.PatchTransactionRequest) bool {
	return req.Type == nil &&
		req.Amount == nil &&
		req.Note == nil &&
		req.TransactionDate == nil &&
		req.CategoryID == nil
}

func ApplyTransactionPatch(
	current model.Transaction,
	req model.PatchTransactionRequest,
) model.Transaction {
	if req.Type != nil {
		current.Type = *req.Type
	}
	if req.Amount != nil {
		current.Amount = *req.Amount
	}
	if req.Note != nil {
		current.Note = *req.Note
	}
	if req.TransactionDate != nil {
		current.TransactionDate = *req.TransactionDate
	}
	if req.CategoryID != nil {
		current.CategoryID = *req.CategoryID
	}

	return current
}

func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}

	return (total + limit - 1) / limit
}

func CategoryMatchesType(categoryType, transactionType string) bool {
	return categoryType != "" && categoryType == transactionType
}
