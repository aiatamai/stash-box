package system

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/stashapp/stash-box/internal/config"
	"github.com/stashapp/stash-box/internal/service"
	"github.com/stashapp/stash-box/pkg/logger"
)

type Handler struct {
	fac     service.Factory
	version string
	hash    string
	stamp   string
}

func NewHandler(fac service.Factory, version, hash, stamp string) *Handler {
	return &Handler{
		fac:     fac,
		version: version,
		hash:    hash,
		stamp:   stamp,
	}
}

// Routes sets up system-related routes
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/version", h.GetVersion)
	r.Get("/config", h.GetConfig)

	return r
}

// VersionResponse represents version information
type VersionResponse struct {
	Version string `json:"version"`
	Hash    string `json:"hash"`
	Stamp   string `json:"stamp"`
}

// ConfigResponse represents instance configuration
type ConfigResponse struct {
	IsProduction bool `json:"is_production"`
}

// @Summary Get version information
// @Tags system
// @Success 200 {object} SuccessResponse{data=VersionResponse}
// @Router /system/version [get]
func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"data": VersionResponse{
			Version: h.version,
			Hash:    h.hash,
			Stamp:   h.stamp,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response:", err)
	}
}

// @Summary Get instance configuration
// @Tags system
// @Success 200 {object} SuccessResponse{data=ConfigResponse}
// @Router /system/config [get]
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"data": ConfigResponse{
			IsProduction: config.GetIsProduction(),
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response:", err)
	}
}

type SuccessResponse struct {
	Data interface{} `json:"data"`
}
