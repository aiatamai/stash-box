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

// Helper function to convert nullable int to int with default
func defaultInt(p *int, defaultVal int) int {
	if p == nil {
		return defaultVal
	}
	return *p
}

// CreateSceneEditRequest contains scene edit input
type CreateSceneEditRequest struct {
	Title        *string          `json:"title,omitempty"`
	Details      *string          `json:"details,omitempty"`
	Date         *string          `json:"date,omitempty"`
	Duration     *int             `json:"duration,omitempty"`
	Director     *string          `json:"director,omitempty"`
	Code         *string          `json:"code,omitempty"`
	StudioID     *string          `json:"studio_id,omitempty"`
	PerformerIDs *[]string        `json:"performer_ids,omitempty"`
	TagIDs       *[]string        `json:"tag_ids,omitempty"`
	URLs         *[]string        `json:"urls,omitempty"`
	Fingerprints *[]FingerprintInput `json:"fingerprints,omitempty"`
}

type FingerprintInput struct {
	Hash      string `json:"hash"`
	Algorithm string `json:"algorithm"`
	Duration  *int   `json:"duration,omitempty"`
}

// UpdateSceneEditRequest contains scene edit update data
type UpdateSceneEditRequest struct {
	Title        *string          `json:"title,omitempty"`
	Details      *string          `json:"details,omitempty"`
	Date         *string          `json:"date,omitempty"`
	Duration     *int             `json:"duration,omitempty"`
	Director     *string          `json:"director,omitempty"`
	Code         *string          `json:"code,omitempty"`
	StudioID     *string          `json:"studio_id,omitempty"`
	PerformerIDs *[]string        `json:"performer_ids,omitempty"`
	TagIDs       *[]string        `json:"tag_ids,omitempty"`
	URLs         *[]string        `json:"urls,omitempty"`
	Fingerprints *[]FingerprintInput `json:"fingerprints,omitempty"`
}

// @Summary Create a new scene edit
// @Tags scenes
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Scene ID (UUID)"
// @Param input body CreateSceneEditRequest true "Scene edit data"
// @Success 200 {object} handlers.SuccessResponse{data=models.Edit}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /scenes/{id}/edits [post]
func (h *Handler) CreateEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	sceneID, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid scene ID format"))
		return
	}

	var req CreateSceneEditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	_ = sceneID // TODO: Pass scene ID to edit creation

	// Build scene edit details
	details := &models.SceneEditDetailsInput{
		Title:      req.Title,
		Details:    req.Details,
		Date:       req.Date,
		Duration:   req.Duration,
		Director:   req.Director,
		Code:       req.Code,
	}

	// Convert studio ID
	if req.StudioID != nil {
		if studioID, err := uuid.FromString(*req.StudioID); err == nil {
			details.StudioID = &studioID
		}
	}

	// Convert performer IDs to appearances
	if req.PerformerIDs != nil {
		performers := make([]models.PerformerAppearanceInput, 0, len(*req.PerformerIDs))
		for _, idStr := range *req.PerformerIDs {
			if performerID, err := uuid.FromString(idStr); err == nil {
				performers = append(performers, models.PerformerAppearanceInput{
					PerformerID: performerID,
				})
			}
		}
		if len(performers) > 0 {
			details.Performers = performers
		}
	}

	// Convert tag IDs
	if req.TagIDs != nil {
		tagIDs := make([]uuid.UUID, 0, len(*req.TagIDs))
		for _, idStr := range *req.TagIDs {
			if tagID, err := uuid.FromString(idStr); err == nil {
				tagIDs = append(tagIDs, tagID)
			}
		}
		if len(tagIDs) > 0 {
			details.TagIds = tagIDs
		}
	}

	// Convert URLs
	if req.URLs != nil {
		urls := make([]models.URL, 0, len(*req.URLs))
		for _, url := range *req.URLs {
			urls = append(urls, models.URL{URL: url})
		}
		if len(urls) > 0 {
			details.Urls = urls
		}
	}

	// Convert fingerprints
	if req.Fingerprints != nil {
		fingerprints := make([]models.FingerprintInput, 0, len(*req.Fingerprints))
		for _, fp := range *req.Fingerprints {
			var algorithm models.FingerprintAlgorithm
			switch fp.Algorithm {
			case "MD5":
				algorithm = models.FingerprintAlgorithmMd5
			case "OSHASH":
				algorithm = models.FingerprintAlgorithmOshash
			case "PHASH":
				algorithm = models.FingerprintAlgorithmPhash
			default:
				continue // Skip invalid algorithms
			}

			fingerprints = append(fingerprints, models.FingerprintInput{
				Hash:      fp.Hash,
				Algorithm: algorithm,
				Duration:  defaultInt(fp.Duration, 0),
			})
		}
		if len(fingerprints) > 0 {
			details.Fingerprints = fingerprints
		}
	}

	// Create the edit
	input := models.SceneEditInput{
		Edit:    &models.EditInput{Operation: models.OperationEnumModify},
		Details: details,
	}

	// Use the service layer method to create scene edit
	edit, err := h.fac.Edit().CreateSceneEdit(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, edit)
}

// @Summary Update a pending scene edit
// @Tags scenes
// @Security ApiKeyAuth SessionAuth
// @Param editId path string true "Edit ID (UUID)"
// @Param input body UpdateSceneEditRequest true "Updated scene edit data"
// @Success 200 {object} handlers.SuccessResponse{data=models.Edit}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /scenes/edits/{editId} [put]
func (h *Handler) UpdateEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	editIDStr := chi.URLParam(r, "editId")

	editID, err := uuid.FromString(editIDStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid edit ID format"))
		return
	}

	var req UpdateSceneEditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	// Build scene edit details
	details := &models.SceneEditDetailsInput{
		Title:      req.Title,
		Details:    req.Details,
		Date:       req.Date,
		Duration:   req.Duration,
		Director:   req.Director,
		Code:       req.Code,
	}

	// Convert studio ID
	if req.StudioID != nil {
		if studioID, err := uuid.FromString(*req.StudioID); err == nil {
			details.StudioID = &studioID
		}
	}

	// Convert performer IDs to appearances
	if req.PerformerIDs != nil {
		performers := make([]models.PerformerAppearanceInput, 0, len(*req.PerformerIDs))
		for _, idStr := range *req.PerformerIDs {
			if performerID, err := uuid.FromString(idStr); err == nil {
				performers = append(performers, models.PerformerAppearanceInput{
					PerformerID: performerID,
				})
			}
		}
		if len(performers) > 0 {
			details.Performers = performers
		}
	}

	// Convert tag IDs
	if req.TagIDs != nil {
		tagIDs := make([]uuid.UUID, 0, len(*req.TagIDs))
		for _, idStr := range *req.TagIDs {
			if tagID, err := uuid.FromString(idStr); err == nil {
				tagIDs = append(tagIDs, tagID)
			}
		}
		if len(tagIDs) > 0 {
			details.TagIds = tagIDs
		}
	}

	// Convert URLs
	if req.URLs != nil {
		urls := make([]models.URL, 0, len(*req.URLs))
		for _, url := range *req.URLs {
			urls = append(urls, models.URL{URL: url})
		}
		if len(urls) > 0 {
			details.Urls = urls
		}
	}

	// Convert fingerprints
	if req.Fingerprints != nil {
		fingerprints := make([]models.FingerprintInput, 0, len(*req.Fingerprints))
		for _, fp := range *req.Fingerprints {
			var algorithm models.FingerprintAlgorithm
			switch fp.Algorithm {
			case "MD5":
				algorithm = models.FingerprintAlgorithmMd5
			case "OSHASH":
				algorithm = models.FingerprintAlgorithmOshash
			case "PHASH":
				algorithm = models.FingerprintAlgorithmPhash
			default:
				continue // Skip invalid algorithms
			}

			fingerprints = append(fingerprints, models.FingerprintInput{
				Hash:      fp.Hash,
				Algorithm: algorithm,
				Duration:  defaultInt(fp.Duration, 0),
			})
		}
		if len(fingerprints) > 0 {
			details.Fingerprints = fingerprints
		}
	}

	// Create the edit input
	input := models.SceneEditInput{
		Edit:    &models.EditInput{Operation: models.OperationEnumModify},
		Details: details,
	}

	// Use the service layer method to update scene edit
	edit, err := h.fac.Edit().UpdateSceneEdit(ctx, editID, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, edit)
}
