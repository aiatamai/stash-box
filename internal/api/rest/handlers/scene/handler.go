package scene

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/go-chi/chi/v5"
	"github.com/stashapp/stash-box/internal/api/rest/handlers"
	"github.com/stashapp/stash-box/internal/models"
	"github.com/stashapp/stash-box/internal/service"
)

type Handler struct {
	fac service.Factory
}

func NewHandler(fac service.Factory) *Handler {
	return &Handler{
		fac: fac,
	}
}

// Routes sets up scene-related routes
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// CRUD routes
	r.Get("/", h.List)           // GET /scenes
	r.Post("/", h.Create)        // POST /scenes
	r.Get("/{id}", h.Get)        // GET /scenes/{id}
	r.Put("/{id}", h.Update)     // PUT /scenes/{id}
	r.Delete("/{id}", h.Delete)  // DELETE /scenes/{id}

	// Search and specialized routes
	r.Get("/search", h.Search)             // GET /scenes/search
	r.Post("/check-existing", h.CheckExisting) // POST /scenes/check-existing

	// Fingerprint routes
	r.Post("/fingerprints/match", h.FingerprintMatch)   // POST /scenes/fingerprints/match
	r.Post("/fingerprints/batch", h.FingerprintBatch)   // POST /scenes/fingerprints/batch
	r.Post("/{id}/fingerprints", h.SubmitFingerprint)   // POST /scenes/{id}/fingerprints

	// Edit routes
	r.Post("/{id}/edits", h.CreateEdit)          // POST /scenes/{id}/edits
	r.Put("/edits/{editId}", h.UpdateEdit)       // PUT /scenes/edits/{editId}

	return r
}

// @Summary List scenes with filters
// @Tags scenes
// @Security ApiKeyAuth SessionAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(25)
// @Param title query string false "Filter by title"
// @Param studio_id query string false "Filter by studio ID"
// @Param performer_ids query string false "Comma-separated performer IDs"
// @Param tag_ids query string false "Comma-separated tag IDs"
// @Param date query string false "Filter by date (YYYY-MM-DD)"
// @Param date_from query string false "Filter by date from (YYYY-MM-DD)"
// @Param date_to query string false "Filter by date to (YYYY-MM-DD)"
// @Param sort query string false "Sort field (date, title, created_at, updated_at)"
// @Param direction query string false "Sort direction (asc, desc)"
// @Success 200 {object} handlers.PaginatedResponse{data=[]models.Scene}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /scenes [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse pagination
	pagination := handlers.ParsePaginationParams(r)

	// Parse query input
	input := models.SceneQueryInput{
		Page:      pagination.Page,
		PerPage:   pagination.PerPage,
		Direction: models.SortDirectionEnum(strings.ToUpper(pagination.Direction)),
	}

	// Parse title filter
	if title := r.URL.Query().Get("title"); title != "" {
		input.Title = &title
	}

	// Parse studio ID filter
	if studioIDStr := r.URL.Query().Get("studio_id"); studioIDStr != "" {
		if studioID, err := uuid.FromString(studioIDStr); err == nil {
			input.Studios = &models.MultiIDCriterionInput{
				Value:    []uuid.UUID{studioID},
				Modifier: models.CriterionModifierIncludes,
			}
		}
	}

	// Parse sort field
	if pagination.Sort != "" {
		input.Sort = models.SceneSortEnum(pagination.Sort)
	}

	// Query scenes
	scenes, err := h.fac.Scene().Query(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Get total count
	count, err := h.fac.Scene().QueryCount(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WritePaginated(w, scenes, pagination.Page, pagination.PerPage, count)
}

// @Summary Get scene by ID
// @Tags scenes
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Scene ID (UUID)"
// @Success 200 {object} handlers.SuccessResponse{data=models.Scene}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /scenes/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid scene ID format"))
		return
	}

	scene, err := h.fac.Scene().FindByID(ctx, id)
	if err != nil {
		handlers.WriteError(w, http.StatusNotFound, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, scene)
}

// @Summary Create a new scene
// @Tags scenes
// @Security ApiKeyAuth SessionAuth
// @Param input body models.SceneCreateInput true "Scene data"
// @Success 201 {object} handlers.SuccessResponse{data=models.Scene}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /scenes [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input models.SceneCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	scene, err := h.fac.Scene().Create(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, scene)
}

// @Summary Update a scene
// @Tags scenes
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Scene ID (UUID)"
// @Param input body models.SceneUpdateInput true "Updated scene data"
// @Success 200 {object} handlers.SuccessResponse{data=models.Scene}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /scenes/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid scene ID format"))
		return
	}

	var input models.SceneUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	input.ID = id
	scene, err := h.fac.Scene().Update(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, scene)
}

// @Summary Delete a scene
// @Tags scenes
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Scene ID (UUID)"
// @Success 204
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /scenes/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid scene ID format"))
		return
	}

	if err := h.fac.Scene().Delete(ctx, id); err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Search scenes by term
// @Tags scenes
// @Security ApiKeyAuth SessionAuth
// @Param term query string true "Search term"
// @Param limit query int false "Result limit" default(25)
// @Success 200 {object} handlers.SuccessResponse{data=[]models.Scene}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /scenes/search [get]
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	term := r.URL.Query().Get("term")
	if term == "" {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("term parameter required"))
		return
	}

	limit := 25
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	scenes, err := h.fac.Scene().SearchScenes(ctx, term, limit)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, scenes)
}

// @Summary Check for existing scenes
// @Tags scenes
// @Security ApiKeyAuth SessionAuth
// @Param input body models.QueryExistingSceneInput true "Query criteria"
// @Success 200 {object} handlers.SuccessResponse{data=[]models.Scene}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /scenes/check-existing [post]
func (h *Handler) CheckExisting(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input models.QueryExistingSceneInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	scenes, err := h.fac.Scene().FindExistingScenes(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, scenes)
}

