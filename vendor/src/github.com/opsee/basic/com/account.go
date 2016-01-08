package com

import (
	"time"
)

type Account struct {
	ID         int       `json:"id" db:"id"`
	CustomerID string    `json:"customer_id" db:"customer_id"`
	Active     bool      `json:"active" db:"active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
