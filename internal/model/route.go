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
	ID            int
	Pattern       string
	BackendURL    string
	Method        HTTPMethod
	CreatedAt     time.Time
	UpdatedAt     *time.Time
	InactivatedAt *time.Time
}
