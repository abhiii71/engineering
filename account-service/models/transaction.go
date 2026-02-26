package model

import "time"

type Transaction struct {
	ID          uint64    `db:"id"`
	AccountID   uint64    `db:"account_id"`
	AmountCents int64     `db:"amount_cents"`
	Kind        string    `db:"kind"` // "credit" or "debit"
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}
