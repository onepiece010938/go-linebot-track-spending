// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package postgres

import (
	"time"
)

type Item struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Name         string    `json:"name"`
	BalanceLimit int32     `json:"balance_limit"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Record struct {
	ID        int64     `json:"id"`
	ItemID    int64     `json:"item_id"`
	Name      string    `json:"name"`
	Price     int32     `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID         int64     `json:"id"`
	Lineid     string    `json:"lineid"`
	Name       string    `json:"name"`
	MonthLimit int32     `json:"month_limit"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}