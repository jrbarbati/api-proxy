package model

import "time"

// Route represents a possible API endpoint to push a call to
type Route struct {
	ID            int        `json:"id"`
	Pattern       string     `json:"pattern"`
	BackendURL    string     `json:"backend_url"`
	Method        string     `json:"method"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	InactivatedAt *time.Time `json:"inactivated_at"`
}

type RouteFilter struct {
	Pattern       string
	Method        string
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
}
