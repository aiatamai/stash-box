package studio

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

// Routes sets up studio-related routes
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// CRUD routes
	r.Get("/", h.List)           // GET /studios
	r.Post("/", h.Create)        // POST /studios
	r.Get("/{id}", h.Get)        // GET /studios/{id}
	r.Put("/{id}", h.Update)     // PUT /studios/{id}
	r.Delete("/{id}", h.Delete)  // DELETE /studios/{id}

	// Search and specialized routes
	r.Get("/search", h.Search)             // GET /studios/search
	r.Get("/by-name/{name}", h.GetByName)  // GET /studios/by-name/{name}

	// Favorite routes
	r.Post("/{id}/favorite", h.Favorite)     // POST /studios/{id}/favorite
	r.Delete("/{id}/favorite", h.Unfavorite) // DELETE /studios/{id}/favorite

	return r
}

// @Summary List studios with filters
// @Tags studios
// @Security ApiKeyAuth SessionAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(25)
// @Param name query string false "Filter by name"
// @Param sort query string false "Sort field (name, created_at, updated_at)"
// @Param direction query string false "Sort direction (asc, desc)"
// @Success 200 {object} handlers.PaginatedResponse{data=[]models.Studio}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /studios [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse pagination
	pagination := handlers.ParsePaginationParams(r)

	// Parse query input
	input := models.StudioQueryInput{
		Page:      pagination.Page,
		PerPage:   pagination.PerPage,
		Direction: models.SortDirectionEnum(strings.ToUpper(pagination.Direction)),
	}

	// Parse name filter
	if name := r.URL.Query().Get("name"); name != "" {
		input.Name = &name
	}

	// Parse sort field
	if pagination.Sort != "" {
		input.Sort = models.StudioSortEnum(pagination.Sort)
	}

	// Query studios
	result, err := h.fac.Studio().Query(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WritePaginated(w, result.Studios, pagination.Page, pagination.PerPage, result.Count)
}

// @Summary Get studio by ID
// @Tags studios
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Studio ID (UUID)"
// @Success 200 {object} handlers.SuccessResponse{data=models.Studio}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /studios/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid studio ID format"))
		return
	}

	studio, err := h.fac.Studio().FindByID(ctx, id)
	if err != nil {
		handlers.WriteError(w, http.StatusNotFound, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, studio)
}

// @Summary Get studio by name
// @Tags studios
// @Security ApiKeyAuth SessionAuth
// @Param name path string true "Studio name"
// @Success 200 {object} handlers.SuccessResponse{data=models.Studio}
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /studios/by-name/{name} [get]
func (h *Handler) GetByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := chi.URLParam(r, "name")

	studio, err := h.fac.Studio().FindByName(ctx, name)
	if err != nil {
		handlers.WriteError(w, http.StatusNotFound, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, studio)
}

// @Summary Create a new studio
// @Tags studios
// @Security ApiKeyAuth SessionAuth
// @Param input body models.StudioCreateInput true "Studio data"
// @Success 201 {object} handlers.SuccessResponse{data=models.Studio}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /studios [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input models.StudioCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	studio, err := h.fac.Studio().Create(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, studio)
}

// @Summary Update a studio
// @Tags studios
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Studio ID (UUID)"
// @Param input body models.StudioUpdateInput true "Updated studio data"
// @Success 200 {object} handlers.SuccessResponse{data=models.Studio}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /studios/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid studio ID format"))
		return
	}

	var input models.StudioUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	input.ID = id
	studio, err := h.fac.Studio().Update(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, studio)
}

// @Summary Delete a studio
// @Tags studios
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Studio ID (UUID)"
// @Success 204
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /studios/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid studio ID format"))
		return
	}

	if err := h.fac.Studio().Delete(ctx, id); err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Search studios by term
// @Tags studios
// @Security ApiKeyAuth SessionAuth
// @Param term query string true "Search term"
// @Param limit query int false "Result limit" default(25)
// @Success 200 {object} handlers.SuccessResponse{data=[]models.Studio}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /studios/search [get]
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

	studios, err := h.fac.Studio().Search(ctx, term, limit)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, studios)
}

// @Summary Favorite a studio
// @Tags studios
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Studio ID (UUID)"
// @Success 200 {object} handlers.SuccessResponse{data=map[string]bool}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /studios/{id}/favorite [post]
func (h *Handler) Favorite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid studio ID format"))
		return
	}

	if err := h.fac.Studio().Favorite(ctx, id, true); err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, map[string]bool{"favorited": true})
}

// @Summary Unfavorite a studio
// @Tags studios
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Studio ID (UUID)"
// @Success 200 {object} handlers.SuccessResponse{data=map[string]bool}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /studios/{id}/favorite [delete]
func (h *Handler) Unfavorite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid studio ID format"))
		return
	}

	if err := h.fac.Studio().Favorite(ctx, id, false); err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, map[string]bool{"favorited": false})
}
