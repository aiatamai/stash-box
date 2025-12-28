package performer

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/go-chi/chi/v5"
	"github.com/stashapp/stash-box/internal/api/rest/handlers"
	"github.com/stashapp/stash-box/internal/models"
)

// CreatePerformerEditRequest contains performer edit input
type CreatePerformerEditRequest struct {
	Name            *string          `json:"name,omitempty"`
	Disambiguation  *string          `json:"disambiguation,omitempty"`
	Aliases         *[]string        `json:"aliases,omitempty"`
	Gender          *string          `json:"gender,omitempty"`
	Urls            *[]string        `json:"urls,omitempty"`
	Birthdate       *string          `json:"birthdate,omitempty"`
	Deathdate       *string          `json:"deathdate,omitempty"`
	Ethnicity       *string          `json:"ethnicity,omitempty"`
	Country         *string          `json:"country,omitempty"`
	EyeColor        *string          `json:"eye_color,omitempty"`
	HairColor       *string          `json:"hair_color,omitempty"`
	Height          *int             `json:"height,omitempty"`
	CupSize         *string          `json:"cup_size,omitempty"`
	BandSize        *int             `json:"band_size,omitempty"`
	WaistSize       *int             `json:"waist_size,omitempty"`
	HipSize         *int             `json:"hip_size,omitempty"`
	BreastType      *string          `json:"breast_type,omitempty"`
	CareerStartYear *int             `json:"career_start_year,omitempty"`
	CareerEndYear   *int             `json:"career_end_year,omitempty"`
}

// UpdatePerformerEditRequest contains performer edit update data
type UpdatePerformerEditRequest struct {
	Name            *string          `json:"name,omitempty"`
	Disambiguation  *string          `json:"disambiguation,omitempty"`
	Aliases         *[]string        `json:"aliases,omitempty"`
	Gender          *string          `json:"gender,omitempty"`
	Urls            *[]string        `json:"urls,omitempty"`
	Birthdate       *string          `json:"birthdate,omitempty"`
	Deathdate       *string          `json:"deathdate,omitempty"`
	Ethnicity       *string          `json:"ethnicity,omitempty"`
	Country         *string          `json:"country,omitempty"`
	EyeColor        *string          `json:"eye_color,omitempty"`
	HairColor       *string          `json:"hair_color,omitempty"`
	Height          *int             `json:"height,omitempty"`
	CupSize         *string          `json:"cup_size,omitempty"`
	BandSize        *int             `json:"band_size,omitempty"`
	WaistSize       *int             `json:"waist_size,omitempty"`
	HipSize         *int             `json:"hip_size,omitempty"`
	BreastType      *string          `json:"breast_type,omitempty"`
	CareerStartYear *int             `json:"career_start_year,omitempty"`
	CareerEndYear   *int             `json:"career_end_year,omitempty"`
}

// @Summary Create a new performer edit
// @Tags performers
// @Security ApiKeyAuth SessionAuth
// @Param id path string true "Performer ID (UUID)"
// @Param input body CreatePerformerEditRequest true "Performer edit data"
// @Success 200 {object} handlers.SuccessResponse{data=models.Edit}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /performers/{id}/edits [post]
func (h *Handler) CreateEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	performerID, err := uuid.FromString(idStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid performer ID format"))
		return
	}

	var req CreatePerformerEditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	_ = performerID // TODO: Pass performer ID to edit creation

	// Build performer edit details
	details := &models.PerformerEditDetailsInput{
		Name:            req.Name,
		Disambiguation:  req.Disambiguation,
		Gender:          mapStringToGenderEnum(req.Gender),
		Birthdate:       req.Birthdate,
		Deathdate:       req.Deathdate,
		Ethnicity:       mapStringToEthnicityEnum(req.Ethnicity),
		Country:         req.Country,
		EyeColor:        mapStringToEyeColorEnum(req.EyeColor),
		HairColor:       mapStringToHairColorEnum(req.HairColor),
		Height:          req.Height,
		CupSize:         req.CupSize,
		BandSize:        req.BandSize,
		WaistSize:       req.WaistSize,
		HipSize:         req.HipSize,
		BreastType:      mapStringToBreastTypeEnum(req.BreastType),
		CareerStartYear: req.CareerStartYear,
		CareerEndYear:   req.CareerEndYear,
	}

	// Convert aliases
	if req.Aliases != nil {
		details.Aliases = *req.Aliases
	}

	// Convert URLs
	if req.Urls != nil {
		urls := make([]models.URL, 0, len(*req.Urls))
		for _, url := range *req.Urls {
			urls = append(urls, models.URL{URL: url})
		}
		details.Urls = urls
	}

	// Create the edit input
	input := models.PerformerEditInput{
		Edit:    &models.EditInput{Operation: models.OperationEnumModify},
		Details: details,
	}

	// Use the service layer method to create performer edit
	edit, err := h.fac.Edit().CreatePerformerEdit(ctx, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, edit)
}

// @Summary Update a pending performer edit
// @Tags performers
// @Security ApiKeyAuth SessionAuth
// @Param editId path string true "Edit ID (UUID)"
// @Param input body UpdatePerformerEditRequest true "Updated performer edit data"
// @Success 200 {object} handlers.SuccessResponse{data=models.Edit}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /performers/edits/{editId} [put]
func (h *Handler) UpdateEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	editIDStr := chi.URLParam(r, "editId")

	editID, err := uuid.FromString(editIDStr)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid edit ID format"))
		return
	}

	var req UpdatePerformerEditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	// Build performer edit details
	details := &models.PerformerEditDetailsInput{
		Name:            req.Name,
		Disambiguation:  req.Disambiguation,
		Gender:          mapStringToGenderEnum(req.Gender),
		Birthdate:       req.Birthdate,
		Deathdate:       req.Deathdate,
		Ethnicity:       mapStringToEthnicityEnum(req.Ethnicity),
		Country:         req.Country,
		EyeColor:        mapStringToEyeColorEnum(req.EyeColor),
		HairColor:       mapStringToHairColorEnum(req.HairColor),
		Height:          req.Height,
		CupSize:         req.CupSize,
		BandSize:        req.BandSize,
		WaistSize:       req.WaistSize,
		HipSize:         req.HipSize,
		BreastType:      mapStringToBreastTypeEnum(req.BreastType),
		CareerStartYear: req.CareerStartYear,
		CareerEndYear:   req.CareerEndYear,
	}

	// Convert aliases
	if req.Aliases != nil {
		details.Aliases = *req.Aliases
	}

	// Convert URLs
	if req.Urls != nil {
		urls := make([]models.URL, 0, len(*req.Urls))
		for _, url := range *req.Urls {
			urls = append(urls, models.URL{URL: url})
		}
		details.Urls = urls
	}

	// Create the edit input
	input := models.PerformerEditInput{
		Edit:    &models.EditInput{Operation: models.OperationEnumModify},
		Details: details,
	}

	// Use the service layer method to update performer edit
	edit, err := h.fac.Edit().UpdatePerformerEdit(ctx, editID, input)
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, edit)
}

// Helper functions to convert string enums to model enums
func mapStringToGenderEnum(s *string) *models.GenderEnum {
	if s == nil {
		return nil
	}
	switch *s {
	case "MALE":
		v := models.GenderEnumMale
		return &v
	case "FEMALE":
		v := models.GenderEnumFemale
		return &v
	case "TRANSGENDER_MALE":
		v := models.GenderEnumTransgenderMale
		return &v
	case "TRANSGENDER_FEMALE":
		v := models.GenderEnumTransgenderFemale
		return &v
	case "INTERSEX":
		v := models.GenderEnumIntersex
		return &v
	case "NON_BINARY":
		v := models.GenderEnumNonBinary
		return &v
	default:
		return nil
	}
}

func mapStringToEthnicityEnum(s *string) *models.EthnicityEnum {
	if s == nil {
		return nil
	}
	switch *s {
	case "CAUCASIAN":
		v := models.EthnicityEnumCaucasian
		return &v
	case "BLACK":
		v := models.EthnicityEnumBlack
		return &v
	case "ASIAN":
		v := models.EthnicityEnumAsian
		return &v
	case "INDIAN":
		v := models.EthnicityEnumIndian
		return &v
	case "LATIN":
		v := models.EthnicityEnumLatin
		return &v
	case "MIDDLE_EASTERN":
		v := models.EthnicityEnumMiddleEastern
		return &v
	case "MIXED":
		v := models.EthnicityEnumMixed
		return &v
	case "OTHER":
		v := models.EthnicityEnumOther
		return &v
	default:
		return nil
	}
}

func mapStringToEyeColorEnum(s *string) *models.EyeColorEnum {
	if s == nil {
		return nil
	}
	switch *s {
	case "BLUE":
		v := models.EyeColorEnumBlue
		return &v
	case "BROWN":
		v := models.EyeColorEnumBrown
		return &v
	case "GREY":
		v := models.EyeColorEnumGrey
		return &v
	case "GREEN":
		v := models.EyeColorEnumGreen
		return &v
	case "HAZEL":
		v := models.EyeColorEnumHazel
		return &v
	case "RED":
		v := models.EyeColorEnumRed
		return &v
	default:
		return nil
	}
}

func mapStringToHairColorEnum(s *string) *models.HairColorEnum {
	if s == nil {
		return nil
	}
	switch *s {
	case "BLONDE":
		v := models.HairColorEnumBlonde
		return &v
	case "BRUNETTE":
		v := models.HairColorEnumBrunette
		return &v
	case "BLACK":
		v := models.HairColorEnumBlack
		return &v
	case "RED":
		v := models.HairColorEnumRed
		return &v
	case "AUBURN":
		v := models.HairColorEnumAuburn
		return &v
	case "GREY":
		v := models.HairColorEnumGrey
		return &v
	case "BALD":
		v := models.HairColorEnumBald
		return &v
	case "VARIOUS":
		v := models.HairColorEnumVarious
		return &v
	case "WHITE":
		v := models.HairColorEnumWhite
		return &v
	case "OTHER":
		v := models.HairColorEnumOther
		return &v
	default:
		return nil
	}
}

func mapStringToBreastTypeEnum(s *string) *models.BreastTypeEnum {
	if s == nil {
		return nil
	}
	switch *s {
	case "NATURAL":
		v := models.BreastTypeEnumNatural
		return &v
	case "FAKE":
		v := models.BreastTypeEnumFake
		return &v
	default:
		return nil
	}
}
