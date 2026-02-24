package model

import "time"

// Org represents a client
type Org struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	InactivatedAt *time.Time `json:"inactivated_at"`
}

func (org Org) GetID() int {
	return org.ID
}
