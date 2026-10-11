package service

import (
	"testing"
	"time"

	"api-buku-kas/app/model"
	"api-buku-kas/helper"
)

func TestCountTotalPages(t *testing.T) {
	cases := []struct{ total, limit, want int }{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{137, 20, 7},
		{5, 0, 0},
	}

	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: harap %d, dapat %d",
				tc.total, tc.limit, tc.want, got)
		}
	}
}

func TestApplyTransactionPatch(t *testing.T) {
	date := time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)
	initial := model.Transaction{
		ID: 1, Type: model.TransactionExpense, Amount: 15000,
		Note: "makan siang", TransactionDate: date, CategoryID: 3, OwnerID: 7,
	}

	newAmount := int64(20000)
	result := ApplyTransactionPatch(initial, model.PatchTransactionRequest{
		Amount: &newAmount,
	})

	if result.Amount != 20000 {
		t.Errorf("amount seharusnya 20000, dapat %d", result.Amount)
	}
	if result.Note != "makan siang" || result.CategoryID != 3 ||
		result.Type != model.TransactionExpense || !result.TransactionDate.Equal(date) {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
	if result.OwnerID != 7 {
		t.Error("owner_id tidak boleh berubah lewat PATCH")
	}
}

func TestIsEmptyTransactionPatch(t *testing.T) {
	if !IsEmptyTransactionPatch(model.PatchTransactionRequest{}) {
		t.Error("PATCH kosong seharusnya terdeteksi kosong")
	}

	empty := ""
	if IsEmptyTransactionPatch(model.PatchTransactionRequest{Note: &empty}) {
		t.Error("note kosong yang DIKIRIM bukan PATCH kosong")
	}
}

func TestCanAccessTransaction(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{
		"admin": {"transaction:read:any"},
		"user":  {},
	})

	owner := model.AuthUser{UserID: 1, Role: "user"}
	other := model.AuthUser{UserID: 2, Role: "user"}
	admin := model.AuthUser{UserID: 9, Role: "admin"}
	ghost := model.AuthUser{UserID: 3, Role: "tidak-dikenal"}

	cases := []struct {
		name    string
		current model.AuthUser
		want    bool
	}{
		{"pemilik selalu boleh", owner, true},
		{"user lain ditolak", other, false},
		{"admin dengan permission :any boleh", admin, true},
		{"role tidak dikenal ditolak (fail closed)", ghost, false},
	}

	for _, tc := range cases {
		got := CanAccessTransaction(tc.current, 1, perms, "transaction:read:any")
		if got != tc.want {
			t.Errorf("%s: harap %v, dapat %v", tc.name, tc.want, got)
		}
	}

	if CanAccessTransaction(other, 1, nil, "transaction:read:any") {
		t.Error("PermissionSet nil harus menolak (fail closed)")
	}
}

func TestCategoryMatchesType(t *testing.T) {
	cases := []struct {
		category, transaction string
		want                  bool
	}{
		{model.TransactionIncome, model.TransactionIncome, true},
		{model.TransactionExpense, model.TransactionExpense, true},
		{model.TransactionIncome, model.TransactionExpense, false},
		{model.TransactionExpense, model.TransactionIncome, false},
		{"", model.TransactionIncome, false},
		{"", "", false},
	}

	for _, tc := range cases {
		got := CategoryMatchesType(tc.category, tc.transaction)
		if got != tc.want {
			t.Errorf("kategori=%q transaksi=%q: harap %v, dapat %v",
				tc.category, tc.transaction, tc.want, got)
		}
	}
}