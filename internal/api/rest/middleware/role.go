package middleware

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/stashapp/stash-box/internal/auth"
	"github.com/stashapp/stash-box/internal/models"
	"github.com/stashapp/stash-box/pkg/logger"
)

// writeError writes an error JSON response
func writeError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":   http.StatusText(statusCode),
		"message": err.Error(),
		"code":    statusCode,
	}

	if encErr := json.NewEncoder(w).Encode(response); encErr != nil {
		logger.Error("Failed to encode error response:", encErr)
	}
}

// RequireRole creates middleware that validates user has the required role
func RequireRole(role models.RoleEnum) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if err := auth.ValidateRole(ctx, role); err != nil {
				writeError(w, http.StatusForbidden, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth ensures user is authenticated (any role)
func RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if auth.GetCurrentUser(ctx) == nil {
				writeError(w, http.StatusUnauthorized, errors.New("authentication required"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
