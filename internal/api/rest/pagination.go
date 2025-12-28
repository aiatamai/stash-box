package rest

import (
	"net/http"
	"strconv"
)

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
