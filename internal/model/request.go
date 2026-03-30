package model

import "time"

// Request represents a request to the API proxy
type Request struct {
	ID         int       `json:"id"`
	RouteID    int       `json:"route_id"`
	Method     string    `json:"method"`
	URL        string    `json:"url"`
	StatusCode int       `json:"status_code"`
	Latency    int64     `json:"latency"`
	CreatedAt  time.Time `json:"created_at"`
}
