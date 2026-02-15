package utils

import "net/http"

// StatusRecorder http response code.
type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

// WriteHeader recording the response code in the header.
func (r *StatusRecorder) WriteHeader(statusCode int) {
	r.Status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
