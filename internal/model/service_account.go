package model

import "time"

type ServiceAccount struct {
	ID            int        `json:"id"`
	OrgID         int        `json:"org_id"`
	Identifier    string     `json:"identifier"`
	ClientID      string     `json:"client_id"`
	ClientSecret  string     `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	InactivatedAt *time.Time `json:"inactivated_at"`
}
