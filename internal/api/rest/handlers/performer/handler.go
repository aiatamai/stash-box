package performer

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

// Routes sets up performer-related routes
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// CRUD routes
	r.Get("/", h.List)           // GET /performers
	r.Post("/", h.Create)        // POST /performers
	r.Get("/{id}", h.Get)        // GET /performers/{id}
	r.Put("/{id}", h.Update)     // PUT /performers/{id}
	r.Delete("/{id}", h.Delete)  // DELETE /performers/{id}

	// Search and specialized routes
	r.Get("/search", h.Search)             // GET /performers/search
	r.Post("/check-existing", h.CheckExisting) // POST /performers/check-existing

	// Favorite routes
	r.Post("/{id}/favorite", h.Favorite)     // POST /performers/{id}/favorite
	r.Delete("/{id}/favorite", h.Unfavorite) // DELETE /performers/{id}/favorite

	// Edit routes
	r.Post("/{id}/edits", h.CreateEdit)          // POST /performers/{id}/edits
	r.Put("/edits/{editId}", h.UpdateEdit)       // PUT /performers/edits/{editId}

	return r
}

// @Summary List performers with filters
// @Tags performers
// @Security ApiKeyAuth SessionAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(25)
// @Param name query string false "Filter by name"
// @Param gender query string false "Filter by gender"
// @Param ethnicity query string false "Filter by ethnicity"
// @Param country query string false "Filter by country"
// @Param eye_color query string false "Filter by eye color"
// @Param hair_color query string false "Filter by hair color"
// @Param height_min query int false "Minimum height in cm"
// @Param height_max query int false "Maximum height in cm"
// @Param age_min query int false "Minimum age"
// @Param age_max query int false "Maximum age"
// @Param sort query string false "Sort field (name, birthdate, scene_count, created_at, updated_at)"
// @Param direction query string false "Sort direction (asc, desc)"
// @Success 200 {object} handlers.PaginatedResponse{data=[]models.Performer}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /performers [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse pagination
	pagination := handlers.ParsePaginationParams(r)

	// Parse query input
	input := models.PerformerQueryInput{
		Page:      pagination.Page,
		PerPage:   pagination.PerPage,
		Direction: models.SortDirectionEnum(strings.ToUpper(pagination.Direction)),
	}

	// Parse name filter
	if name := r.URL.Query().Get("name"); name != "" {
		input.Names = &name
	}

	// Parse gender filter
	if gender := r.URL.Query().Get("gender"); gender != "" {
		// Map string to GenderFilterEnum
		switch gender {
		case "MALE", "FEMALE", "TRANSGENDER_MALE", "TRANSGENDER_FEMALE", "INTERSEX", "NON_BINARY", "UNKNOWN":
			genderEnum := models.GenderFilterEnum(gender)
			input.Gender = &genderEnum
		}
	}

	// Parse ethnicity filter
	if ethnicity := r.URL.Query().Get("ethnicity"); ethnicity != "" {
		switch ethnicity {
		case "CAUCASIAN", "BLACK", "ASIAN", "INDIAN", "LATIN", "MIDDLE_EASTERN", "MIXED", "OTHER", "UNKNOWN":
			ethnicityEnum := models.EthnicityFilterEnum(ethnicity)
			input.Ethnicity = &ethnicityEnum
		}
	}

	// Parse country filter
	if country := r.URL.Query().Get("country"); country != "" {
		input.Country = &models.StringCriterionInput{
			Value:      country,
			Modifier:   models.CriterionModifierIncludes,
		}
	}

	// Parse height filters
	if heightMin := r.URL.Query().Get("height_min"); heightMin != "" {
		if h, err := strconv.Atoi(heightMin); err == nil {
			input.Height = &models.IntCriterionInput{
				Value:    h,
				Modifier: models.CriterionModifierGreaterThan,
			}
		}
	}

	// Parse sort field
	if pagination.Sort != "" {
		input.Sort = models.PerformerSortEnum(pagination.Sort)
	}

	// Query performers
	performers, err := h.fac.Performer().Query(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Get total count
	count, err := h.fac.Performer().QueryCount(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WritePaginated(w, performers, pagination.Page, pagination.PerPage, count)
}

// @Summary Get performer by ID
// @Tags performers
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Performer ID (UUID)"
// @Success 200 {object} handlers.SuccessResponse{data=models.Performer}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /performers/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid performer ID format"))
		return
	}

	performer, err := h.fac.Performer().FindByID(ctx, id)
	if err != nil {
		handlers.WriteError(w, http.StatusNotFound, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, performer)
}

// @Summary Create a new performer
// @Tags performers
// @Security ApiKeyAuth SessionAuth
// @Param input body models.PerformerCreateInput true "Performer data"
// @Success 201 {object} handlers.SuccessResponse{data=models.Performer}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /performers [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input models.PerformerCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	performer, err := h.fac.Performer().Create(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, performer)
}

// @Summary Update a performer
// @Tags performers
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Performer ID (UUID)"
// @Param input body models.PerformerUpdateInput true "Updated performer data"
// @Success 200 {object} handlers.SuccessResponse{data=models.Performer}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /performers/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid performer ID format"))
		return
	}

	var input models.PerformerUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	input.ID = id
	performer, err := h.fac.Performer().Update(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, performer)
}

// @Summary Delete a performer
// @Tags performers
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Performer ID (UUID)"
// @Success 204
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /performers/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid performer ID format"))
		return
	}

	if err := h.fac.Performer().Delete(ctx, id); err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Search performers by term
// @Tags performers
// @Security ApiKeyAuth SessionAuth
// @Param term query string true "Search term"
// @Param limit query int false "Result limit" default(25)
// @Success 200 {object} handlers.SuccessResponse{data=[]models.Performer}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /performers/search [get]
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

	performers, err := h.fac.Performer().SearchPerformer(ctx, term, &limit)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, performers)
}

// @Summary Check for existing performers
// @Tags performers
// @Security ApiKeyAuth SessionAuth
// @Param input body models.QueryExistingPerformerInput true "Query criteria"
// @Success 200 {object} handlers.SuccessResponse{data=[]models.Performer}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /performers/check-existing [post]
func (h *Handler) CheckExisting(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input models.QueryExistingPerformerInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	performers, err := h.fac.Performer().FindExistingPerformers(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, performers)
}

// @Summary Favorite a performer
// @Tags performers
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Performer ID (UUID)"
// @Success 200 {object} handlers.SuccessResponse{data=map[string]bool}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /performers/{id}/favorite [post]
func (h *Handler) Favorite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid performer ID format"))
		return
	}

	// Get current user from context
	user := ctx.Value("user")
	if user == nil {
		handlers.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	userID := user.(uuid.UUID)

	if err := h.fac.Performer().Favorite(ctx, userID, id, true); err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, map[string]bool{"favorited": true})
}

// @Summary Unfavorite a performer
// @Tags performers
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Performer ID (UUID)"
// @Success 200 {object} handlers.SuccessResponse{data=map[string]bool}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /performers/{id}/favorite [delete]
func (h *Handler) Unfavorite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid performer ID format"))
		return
	}

	// Get current user from context
	user := ctx.Value("user")
	if user == nil {
		handlers.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	userID := user.(uuid.UUID)

	if err := h.fac.Performer().Favorite(ctx, userID, id, false); err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, map[string]bool{"favorited": false})
}
