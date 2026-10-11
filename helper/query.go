package helper

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-buku-kas/app/model"
)

var AppLocation = time.FixedZone("WIB", 7*60*60)

var transactionSortable = map[string]bool{
	"transaction_date": true,
	"amount":           true,
	"created_at":       true,
	"id":               true,
}

const maxPageLimit = 100

func ParseTransactionQuery(c *fiber.Ctx) (model.TransactionQuery, error) {
	q := model.TransactionQuery{
		Page:   1,
		Limit:  10,
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   "transaction_date",
		Order:  "desc",
	}

	if raw := c.Query("page"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return q, BadRequest("page harus berupa angka positif")
		}
		q.Page = value
	}

	if raw := c.Query("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return q, BadRequest("limit harus berupa angka positif")
		}
		if value > maxPageLimit {
			value = maxPageLimit
		}
		q.Limit = value
	}

	if raw := c.Query("type"); raw != "" {
		if raw != model.TransactionIncome && raw != model.TransactionExpense {
			return q, BadRequest("type harus income atau expense")
		}
		q.Type = raw
	}

	if raw := c.Query("category_id"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return q, BadRequest("category_id harus berupa angka positif")
		}
		q.CategoryID = &value
	}

	if raw := c.Query("sort"); raw != "" {
		if !transactionSortable[raw] {
			return q, BadRequest(
				"sort harus salah satu dari: transaction_date, amount, created_at, id",
			)
		}
		q.Sort = raw
	}

	if raw := c.Query("order"); raw != "" {
		order := strings.ToLower(raw)
		if order != "asc" && order != "desc" {
			return q, BadRequest("order harus asc atau desc")
		}
		q.Order = order
	}

	month := c.Query("month")
	fromRaw := c.Query("from")
	toRaw := c.Query("to")

	if month != "" && (fromRaw != "" || toRaw != "") {
		return q, BadRequest("month tidak dapat digabung dengan from atau to")
	}

	if month != "" {
		start, err := time.ParseInLocation("2006-01", month, AppLocation)
		if err != nil {
			return q, BadRequest("month harus berformat YYYY-MM")
		}
		end := start.AddDate(0, 1, 0)
		q.From = &start
		q.To = &end
	}

	if fromRaw != "" {
		value, err := parseDateParam(fromRaw, false)
		if err != nil {
			return q, BadRequest("from harus berformat YYYY-MM-DD atau RFC3339")
		}
		q.From = &value
	}

	if toRaw != "" {
		value, err := parseDateParam(toRaw, true)
		if err != nil {
			return q, BadRequest("to harus berformat YYYY-MM-DD atau RFC3339")
		}
		q.To = &value
	}

	if q.From != nil && q.To != nil && !q.To.After(*q.From) {
		return q, BadRequest("to harus setelah from")
	}

	return q, nil
}

func parseDateParam(raw string, endOfDay bool) (time.Time, error) {
	if value, err := time.Parse(time.RFC3339, raw); err == nil {
		return value, nil
	}

	day, err := time.ParseInLocation("2006-01-02", raw, AppLocation)
	if err != nil {
		return time.Time{}, err
	}

	if endOfDay {
		day = day.AddDate(0, 0, 1)
	}

	return day, nil
}