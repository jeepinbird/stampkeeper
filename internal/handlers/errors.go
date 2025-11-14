package handlers

import (
	"log"
	"net/http"
)

// LogAndReturnError logs detailed error internally and returns generic message to client
func LogAndReturnError(w http.ResponseWriter, err error, context string, statusCode int) {
	log.Printf("ERROR [%s]: %v", context, err)
	http.Error(w, getGenericMessage(statusCode), statusCode)
}

// getGenericMessage returns a user-friendly generic message for each status code
func getGenericMessage(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "Invalid request"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusForbidden:
		return "Forbidden"
	case http.StatusNotFound:
		return "Resource not found"
	case http.StatusConflict:
		return "Resource conflict"
	case http.StatusInternalServerError:
		return "Internal server error"
	case http.StatusServiceUnavailable:
		return "Service temporarily unavailable"
	default:
		return "An error occurred"
	}
}
