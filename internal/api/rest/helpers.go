package rest

import (
	"encoding/json"
	"net/http"

	"github.com/stashapp/stash-box/pkg/logger"
)

// SuccessResponse wraps successful response data
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// ErrorResponse standardized error structure
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Code    int               `json:"code"`
	Details map[string]string `json:"details,omitempty"`
}

// PaginatedResponse wraps paginated list responses
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalCount int         `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}

// WriteJSON writes a successful JSON response
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := SuccessResponse{
		Data: data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response:", err)
	}
}

// WriteError writes an error JSON response
func WriteError(w http.ResponseWriter, statusCode int, err error, details ...map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorMsg := err.Error()

	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: errorMsg,
		Code:    statusCode,
	}

	if len(details) > 0 {
		response.Details = details[0]
	}

	if encErr := json.NewEncoder(w).Encode(response); encErr != nil {
		logger.Error("Failed to encode error response:", encErr)
	}
}

// WritePaginated writes a paginated JSON response
func WritePaginated(w http.ResponseWriter, data interface{}, page, perPage, totalCount int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	totalPages := (totalCount + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	response := PaginatedResponse{
		Data:       data,
		Page:       page,
		PerPage:    perPage,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode paginated response:", err)
	}
}
