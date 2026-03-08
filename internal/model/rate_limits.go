package model

import "time"

type RateLimit struct {
	ID               int        `json:"id"`
	OrgID            int        `json:"org_id"`
	ServiceAccountID *int       `json:"service_account_id"`
	LimitPerMinute   int        `json:"limit_per_minute"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
	InactivatedAt    *time.Time `json:"inactivated_at"`
}

type RateLimitFilter struct {
	OrgId            *int
	ServiceAccountId *int
}
