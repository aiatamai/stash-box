package tag

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/go-chi/chi/v5"
	"github.com/stashapp/stash-box/internal/api/rest/handlers"
	"github.com/stashapp/stash-box/internal/models"
)

// @Summary List all tag categories
// @Tags tag-categories
// @Security ApiKeyAuth SessionAuth
// @Success 200 {object} handlers.SuccessResponse{data=[]models.TagCategory}
// @Failure 401 {object} handlers.ErrorResponse
// @Router /tags/categories [get]
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	count, categories, err := h.fac.Tag().QueryCategories(ctx)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WritePaginated(w, categories, 1, count, count)
}

// @Summary Get tag category by ID
// @Tags tag-categories
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Category ID (UUID)"
// @Success 200 {object} handlers.SuccessResponse{data=models.TagCategory}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /tags/categories/{id} [get]
func (h *Handler) GetCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid category ID format"))
		return
	}

	category, err := h.fac.Tag().FindCategory(ctx, id)
	if err != nil {
		handlers.WriteError(w, http.StatusNotFound, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, category)
}

// @Summary Create a new tag category
// @Tags tag-categories
// @Security ApiKeyAuth SessionAuth
// @Param input body models.TagCategoryCreateInput true "Category data"
// @Success 201 {object} handlers.SuccessResponse{data=models.TagCategory}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /tags/categories [post]
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input models.TagCategoryCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	category, err := h.fac.Tag().CreateCategory(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, category)
}

// @Summary Update a tag category
// @Tags tag-categories
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Category ID (UUID)"
// @Param input body models.TagCategoryUpdateInput true "Updated category data"
// @Success 200 {object} handlers.SuccessResponse{data=models.TagCategory}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /tags/categories/{id} [put]
func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid category ID format"))
		return
	}

	var input models.TagCategoryUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	input.ID = id
	category, err := h.fac.Tag().UpdateCategory(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, category)
}

// @Summary Delete a tag category
// @Tags tag-categories
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Category ID (UUID)"
// @Success 204
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Router /tags/categories/{id} [delete]
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid category ID format"))
		return
	}

	if err := h.fac.Tag().DeleteCategory(ctx, models.TagCategoryDestroyInput{ID: id}); err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
