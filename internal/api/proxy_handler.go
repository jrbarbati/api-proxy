package api

import (
	"net/http"
)

type ProxyHandler struct {
}

func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{}
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	
}
