package api

import (
	"api-proxy/internal/api/middleware"
	"io"
	"log"
	"net/http"
)

type ProxyHandler struct {
}

func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{}
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	matchedRoute := middleware.MatchedRoute(r)
	url := matchedRoute.BackendURL + r.URL.RequestURI()
	body := r.Body
	defer body.Close()

	request, err := http.NewRequest(r.Method, url, body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	request.Header = r.Header

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	err = writeResponse(w, response)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func writeResponse(w http.ResponseWriter, response *http.Response) error {
	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		return err
	}

	for key, values := range response.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}

	}
	w.WriteHeader(response.StatusCode)

	if _, err := w.Write(responseBody); err != nil {
		log.Printf("error writing response body: %v", err)
	}

	return nil
}
