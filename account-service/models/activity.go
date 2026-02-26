package model

import "time"

type ActivityLog struct {
	ID        uint64    `db:"id"`
	AccountID uint64    `db:"account_id"`
	Action    string    `db:"action"`
	IPAddress string    `db:"ip_address"`
	CreatedAt time.Time `db:"created_at"`
}
