package scene

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/go-chi/chi/v5"
	"github.com/stashapp/stash-box/internal/api/rest/handlers"
	"github.com/stashapp/stash-box/internal/models"
)

// FingerprintQuery represents a single fingerprint for matching
type FingerprintQuery struct {
	Hash      string `json:"hash" binding:"required"`
	Algorithm string `json:"algorithm" binding:"required"`
}

// FingerprintMatchRequest contains a single fingerprint to match
type FingerprintMatchRequest struct {
	Hash      string `json:"hash" binding:"required"`
	Algorithm string `json:"algorithm" binding:"required"`
}

// FingerprintBatchRequest contains multiple scene fingerprints for batch matching
type FingerprintBatchRequest struct {
	Fingerprints [][]FingerprintQuery `json:"fingerprints" binding:"required"`
}

// SubmitFingerprintRequest contains fingerprint submission data
type SubmitFingerprintRequest struct {
	Hash      string `json:"hash" binding:"required"`
	Algorithm string `json:"algorithm" binding:"required"`
	Duration  *int   `json:"duration,omitempty"`
}

// @Summary Match a single fingerprint
// @Tags scenes
// @Security ApiKeyAuth SessionAuth
// @Param input body FingerprintMatchRequest true "Fingerprint data"
// @Success 200 {object} handlers.SuccessResponse{data=[]models.Scene}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /scenes/fingerprints/match [post]
func (h *Handler) FingerprintMatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req FingerprintMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	// Parse algorithm
	var algorithm models.FingerprintAlgorithm
	switch req.Algorithm {
	case "MD5":
		algorithm = models.FingerprintAlgorithmMd5
	case "OSHASH":
		algorithm = models.FingerprintAlgorithmOshash
	case "PHASH":
		algorithm = models.FingerprintAlgorithmPhash
	default:
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid fingerprint algorithm"))
		return
	}

	// Find scenes by fingerprint
	scenes, err := h.fac.Scene().FindByFingerprint(ctx, algorithm, req.Hash)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, scenes)
}

// @Summary Match multiple fingerprints in batch (Plex critical)
// @Tags scenes
// @Security ApiKeyAuth SessionAuth
// @Param input body FingerprintBatchRequest true "Array of fingerprint arrays (one array per scene)"
// @Success 200 {object} handlers.SuccessResponse{data=[][]models.Scene}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 413 {object} handlers.ErrorResponse
// @Router /scenes/fingerprints/batch [post]
func (h *Handler) FingerprintBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req FingerprintBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	// Validate batch size (limit to 40 scenes like GraphQL)
	if len(req.Fingerprints) > 40 {
		handlers.WriteError(w, http.StatusRequestEntityTooLarge, errors.New("too many scenes (max 40)"))
		return
	}

	// Convert to models.FingerprintQueryInput format
	sceneFingerprints := make([][]models.FingerprintQueryInput, len(req.Fingerprints))
	for i, fpQueries := range req.Fingerprints {
		fpInputs := make([]models.FingerprintQueryInput, len(fpQueries))
		for j, fp := range fpQueries {
			var algorithm models.FingerprintAlgorithm
			switch fp.Algorithm {
			case "MD5":
				algorithm = models.FingerprintAlgorithmMd5
			case "OSHASH":
				algorithm = models.FingerprintAlgorithmOshash
			case "PHASH":
				algorithm = models.FingerprintAlgorithmPhash
			default:
				handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid fingerprint algorithm in batch"))
				return
			}

			fpInputs[j] = models.FingerprintQueryInput{
				Hash:      fp.Hash,
				Algorithm: algorithm,
			}
		}
		sceneFingerprints[i] = fpInputs
	}

	// Call service layer
	results, err := h.fac.Scene().FindScenesBySceneFingerprints(ctx, sceneFingerprints)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, results)
}

// @Summary Submit a fingerprint for a scene
// @Tags scenes
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Scene ID (UUID)"
// @Param input body SubmitFingerprintRequest true "Fingerprint to submit"
// @Success 200 {object} handlers.SuccessResponse{data=map[string]bool}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /scenes/{id}/fingerprints [post]
func (h *Handler) SubmitFingerprint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid scene ID format"))
		return
	}

	var req SubmitFingerprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	// Parse algorithm
	var algorithm models.FingerprintAlgorithm
	switch req.Algorithm {
	case "MD5":
		algorithm = models.FingerprintAlgorithmMd5
	case "OSHASH":
		algorithm = models.FingerprintAlgorithmOshash
	case "PHASH":
		algorithm = models.FingerprintAlgorithmPhash
	default:
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid fingerprint algorithm"))
		return
	}

	// Create submission
	duration := 0
	if req.Duration != nil {
		duration = *req.Duration
	}
	input := models.FingerprintSubmission{
		SceneID: id,
		Fingerprint: &models.FingerprintInput{
			Hash:      req.Hash,
			Algorithm: algorithm,
			Duration:  duration,
		},
	}

	// Submit fingerprint
	success, err := h.fac.Scene().SubmitFingerprint(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]bool{
		"submitted": success,
	}

	handlers.WriteJSON(w, http.StatusOK, response)
}
