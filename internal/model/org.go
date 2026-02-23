package model

import "time"

// Org represents a client
type Org struct {
	ID            int
	Name          string
	CreatedAt     time.Time
	UpdatedAt     *time.Time
	InactivatedAt *time.Time
}
