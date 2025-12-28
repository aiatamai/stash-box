package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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

// PaginationParams contains pagination and sorting parameters
type PaginationParams struct {
	Page      int
	PerPage   int
	Sort      string
	Direction string
}

// DefaultPageSize is the default number of items per page
const DefaultPageSize = 25

// MaxPageSize is the maximum number of items per page
const MaxPageSize = 100

// ParsePaginationParams extracts pagination parameters from query string
func ParsePaginationParams(r *http.Request) PaginationParams {
	params := PaginationParams{
		Page:      1,
		PerPage:   DefaultPageSize,
		Sort:      "",
		Direction: "asc",
	}

	// Parse page
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	// Parse per_page with max limit
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 {
			if perPage > MaxPageSize {
				perPage = MaxPageSize
			}
			params.PerPage = perPage
		}
	}

	// Parse sort field
	if sort := r.URL.Query().Get("sort"); sort != "" {
		params.Sort = sort
	}

	// Parse sort direction
	if direction := r.URL.Query().Get("direction"); direction == "desc" || direction == "DESC" {
		params.Direction = "desc"
	}

	return params
}

// GetOffset calculates the database offset from page and per_page
func (p PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PerPage
}

// GetLimit returns the per_page value (limit for database query)
func (p PaginationParams) GetLimit() int {
	return p.PerPage
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
