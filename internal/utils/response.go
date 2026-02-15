package utils

import (
	"encoding/json"
	"net/http"
)

// Response structure for the answer.
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SendJSONResponse sends a JSON response.
func SendJSONResponse(w http.ResponseWriter, success bool, message string, data interface{}) {
	response := Response{
		Success: success,
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}
