package tag

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

// Routes sets up tag-related routes
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// Tag CRUD routes
	r.Get("/", h.List)           // GET /tags
	r.Post("/", h.Create)        // POST /tags
	r.Get("/{id}", h.Get)        // GET /tags/{id}
	r.Put("/{id}", h.Update)     // PUT /tags/{id}
	r.Delete("/{id}", h.Delete)  // DELETE /tags/{id}

	// Tag search and specialized routes
	r.Get("/search", h.Search)             // GET /tags/search
	r.Get("/by-name/{name}", h.GetByName)  // GET /tags/by-name/{name}
	r.Get("/by-alias/{alias}", h.GetByAlias) // GET /tags/by-alias/{alias}

	// Tag category routes
	r.Get("/categories", h.ListCategories)       // GET /tags/categories
	r.Post("/categories", h.CreateCategory)      // POST /tags/categories
	r.Get("/categories/{id}", h.GetCategory)     // GET /tags/categories/{id}
	r.Put("/categories/{id}", h.UpdateCategory)  // PUT /tags/categories/{id}
	r.Delete("/categories/{id}", h.DeleteCategory) // DELETE /tags/categories/{id}

	return r
}

// @Summary List tags with filters
// @Tags tags
// @Security ApiKeyAuth SessionAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(25)
// @Param name query string false "Filter by name"
// @Param category_id query string false "Filter by category ID"
// @Param sort query string false "Sort field (name, created_at, updated_at)"
// @Param direction query string false "Sort direction (asc, desc)"
// @Success 200 {object} handlers.PaginatedResponse{data=[]models.Tag}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /tags [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse pagination
	pagination := handlers.ParsePaginationParams(r)

	// Parse query input
	input := models.TagQueryInput{
		Page:      pagination.Page,
		PerPage:   pagination.PerPage,
		Direction: models.SortDirectionEnum(strings.ToUpper(pagination.Direction)),
	}

	// Parse name filter
	if name := r.URL.Query().Get("name"); name != "" {
		input.Name = &name
	}

	// Parse category ID filter
	if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
		if categoryID, err := uuid.FromString(categoryIDStr); err == nil {
			input.CategoryID = &categoryID
		}
	}

	// Parse sort field
	if pagination.Sort != "" {
		input.Sort = models.TagSortEnum(pagination.Sort)
	}

	// Query tags
	result, err := h.fac.Tag().Query(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WritePaginated(w, result.Tags, pagination.Page, pagination.PerPage, result.Count)
}

// @Summary Get tag by ID
// @Tags tags
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Tag ID (UUID)"
// @Success 200 {object} handlers.SuccessResponse{data=models.Tag}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /tags/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid tag ID format"))
		return
	}

	tag, err := h.fac.Tag().Find(ctx, id)
	if err != nil {
		handlers.WriteError(w, http.StatusNotFound, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, tag)
}

// @Summary Get tag by name
// @Tags tags
// @Security ApiKeyAuth SessionAuth
// @Param name path string true "Tag name"
// @Success 200 {object} handlers.SuccessResponse{data=models.Tag}
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /tags/by-name/{name} [get]
func (h *Handler) GetByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := chi.URLParam(r, "name")

	tag, err := h.fac.Tag().FindByName(ctx, name)
	if err != nil {
		handlers.WriteError(w, http.StatusNotFound, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, tag)
}

// @Summary Get tag by alias
// @Tags tags
// @Security ApiKeyAuth SessionAuth
// @Param alias path string true "Tag alias"
// @Success 200 {object} handlers.SuccessResponse{data=models.Tag}
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /tags/by-alias/{alias} [get]
func (h *Handler) GetByAlias(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	alias := chi.URLParam(r, "alias")

	tag, err := h.fac.Tag().FindByAlias(ctx, alias)
	if err != nil {
		handlers.WriteError(w, http.StatusNotFound, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, tag)
}

// @Summary Create a new tag
// @Tags tags
// @Security ApiKeyAuth SessionAuth
// @Param input body models.TagCreateInput true "Tag data"
// @Success 201 {object} handlers.SuccessResponse{data=models.Tag}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /tags [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input models.TagCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	tag, err := h.fac.Tag().Create(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, tag)
}

// @Summary Update a tag
// @Tags tags
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Tag ID (UUID)"
// @Param input body models.TagUpdateInput true "Updated tag data"
// @Success 200 {object} handlers.SuccessResponse{data=models.Tag}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /tags/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid tag ID format"))
		return
	}

	var input models.TagUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	input.ID = id
	tag, err := h.fac.Tag().Update(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, tag)
}

// @Summary Delete a tag
// @Tags tags
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Tag ID (UUID)"
// @Success 204
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /tags/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid tag ID format"))
		return
	}

	if err := h.fac.Tag().Delete(ctx, models.TagDestroyInput{ID: id}); err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Search tags by term
// @Tags tags
// @Security ApiKeyAuth SessionAuth
// @Param term query string true "Search term"
// @Param limit query int false "Result limit" default(25)
// @Success 200 {object} handlers.SuccessResponse{data=[]models.Tag}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /tags/search [get]
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

	tags, err := h.fac.Tag().SearchTags(ctx, term, limit)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, tags)
}
