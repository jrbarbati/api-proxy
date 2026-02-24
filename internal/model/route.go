package model

import "time"

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	PATCH  HTTPMethod = "PATCH"
	DELETE HTTPMethod = "DELETE"
)

type HTTPMethod string

// Route represents a possible API endpoint to push a call to
type Route struct {
	ID            int        `json:"id"`
	Pattern       string     `json:"pattern"`
	BackendURL    string     `json:"backend_url"`
	Method        HTTPMethod `json:"method"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	InactivatedAt *time.Time `json:"inactivated_at"`
}
