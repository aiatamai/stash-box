package edit

import (
	"encoding/json"
	"errors"
	"net/http"
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

// Routes sets up edit-related routes
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// Query routes
	r.Get("/", h.List)           // GET /edits
	r.Get("/{id}", h.Get)        // GET /edits/{id}

	// Vote routes
	r.Post("/{id}/vote", h.Vote)         // POST /edits/{id}/vote

	// Comment routes
	r.Post("/{id}/comments", h.Comment)  // POST /edits/{id}/comments

	// Management routes
	r.Post("/{id}/apply", h.Apply)       // POST /edits/{id}/apply
	r.Post("/{id}/cancel", h.Cancel)     // POST /edits/{id}/cancel

	return r
}

// @Summary List edits with filters
// @Tags edits
// @Security ApiKeyAuth SessionAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(25)
// @Param status query string false "Filter by status (PENDING, ACCEPTED, REJECTED, etc.)"
// @Param operation query string false "Filter by operation (CREATE, MODIFY, MERGE, DESTROY)"
// @Param target_type query string false "Filter by target type (SCENE, STUDIO, PERFORMER, TAG)"
// @Param user_id query string false "Filter by creator user ID"
// @Param sort query string false "Sort field (created_at, updated_at, closed_at)"
// @Param direction query string false "Sort direction (asc, desc)"
// @Success 200 {object} handlers.PaginatedResponse{data=[]models.Edit}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /edits [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse pagination
	pagination := handlers.ParsePaginationParams(r)

	// Parse query input
	input := models.EditQueryInput{
		Page:      pagination.Page,
		PerPage:   pagination.PerPage,
		Direction: models.SortDirectionEnum(strings.ToUpper(pagination.Direction)),
	}

	// Parse status filter
	if status := r.URL.Query().Get("status"); status != "" {
		statusEnum := models.VoteStatusEnum(status)
		input.Status = &statusEnum
	}

	// Parse operation filter
	if operation := r.URL.Query().Get("operation"); operation != "" {
		opEnum := models.OperationEnum(operation)
		input.Operation = &opEnum
	}

	// Parse target type filter
	if targetType := r.URL.Query().Get("target_type"); targetType != "" {
		targetEnum := models.TargetTypeEnum(targetType)
		input.TargetType = &targetEnum
	}

	// Parse user ID filter
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if userID, err := uuid.FromString(userIDStr); err == nil {
			input.UserID = &userID
		}
	}

	// Parse sort field
	if pagination.Sort != "" {
		input.Sort = models.EditSortEnum(pagination.Sort)
	}

	// Query edits
	edits, err := h.fac.Edit().QueryEdits(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Get count
	count, err := h.fac.Edit().QueryCount(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WritePaginated(w, edits, pagination.Page, pagination.PerPage, count)
}

// @Summary Get edit by ID
// @Tags edits
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Edit ID (UUID)"
// @Success 200 {object} handlers.SuccessResponse{data=models.Edit}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /edits/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid edit ID format"))
		return
	}

	edit, err := h.fac.Edit().FindByID(ctx, id)
	if err != nil {
		handlers.WriteError(w, http.StatusNotFound, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, edit)
}

// VoteRequest contains vote input
type VoteRequest struct {
	Vote string `json:"vote"` // ACCEPT, REJECT, ABSTAIN, IMMEDIATE_ACCEPT, IMMEDIATE_REJECT
}

// @Summary Vote on an edit
// @Tags edits
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Edit ID (UUID)"
// @Param input body VoteRequest true "Vote data"
// @Success 200 {object} handlers.SuccessResponse{data=models.Edit}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 409 {object} handlers.ErrorResponse
// @Router /edits/{id}/vote [post]
func (h *Handler) Vote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid edit ID format"))
		return
	}

	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	// Parse vote type
	voteEnum := models.VoteTypeEnum(req.Vote)
	input := models.EditVoteInput{
		ID:   id,
		Vote: voteEnum,
	}

	edit, err := h.fac.Edit().CreateVote(ctx, input)
	if err != nil {
		// Return 409 Conflict if vote already exists or edit not pending
		handlers.WriteError(w, http.StatusConflict, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, edit)
}

// CommentRequest contains comment input
type CommentRequest struct {
	Comment string `json:"comment"` // Comment text
}

// @Summary Comment on an edit
// @Tags edits
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Edit ID (UUID)"
// @Param input body CommentRequest true "Comment data"
// @Success 201 {object} handlers.SuccessResponse{data=models.Edit}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /edits/{id}/comments [post]
func (h *Handler) Comment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid edit ID format"))
		return
	}

	var req CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	if req.Comment == "" {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("comment cannot be empty"))
		return
	}

	input := models.EditCommentInput{
		ID:      id,
		Comment: req.Comment,
	}

	edit, _, err := h.fac.Edit().CreateComment(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, edit)
}

// @Summary Apply an edit
// @Tags edits
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Edit ID (UUID)"
// @Success 200 {object} handlers.SuccessResponse{data=models.Edit}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /edits/{id}/apply [post]
func (h *Handler) Apply(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid edit ID format"))
		return
	}

	input := models.ApplyEditInput{
		ID: id,
	}

	edit, err := h.fac.Edit().Apply(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, edit)
}

// @Summary Cancel an edit
// @Tags edits
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Edit ID (UUID)"
// @Success 200 {object} handlers.SuccessResponse{data=models.Edit}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /edits/{id}/cancel [post]
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid edit ID format"))
		return
	}

	input := models.CancelEditInput{
		ID: id,
	}

	edit, err := h.fac.Edit().Cancel(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, edit)
}
