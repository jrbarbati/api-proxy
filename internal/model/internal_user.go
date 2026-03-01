package model

import "time"

type InternalUser struct {
	ID            int        `json:"id"`
	Email         string     `json:"email"`
	Password      string     `json:"password,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	InactivatedAt *time.Time `json:"inactivated_at"`
}
