package controller

import (
	"net/http"
	"strings"
	"time"

	"golang-backend/dto"
	"golang-backend/service"
	"golang-backend/utils"

	"github.com/gin-gonic/gin"
)

type CandidateController struct {
	service service.CandidateService
}

func NewCandidateController(candidateService service.CandidateService) *CandidateController {
	return &CandidateController{service: candidateService}
}

// Login godoc
// @Summary Candidate login
// @Tags candidate-auth
// @Accept json
// @Produce json
// @Param input body dto.CandidateLoginRequest true "Candidate credentials"
// @Success 200 {object} utils.Response
// @Router /v1/auth/candidate/login [post]
func (c *CandidateController) Login(ctx *gin.Context) {
	var request dto.CandidateLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}
	token, candidate, err := c.service.Login(request)
	if err != nil {
		utils.ErrorResponse(ctx, "Login failed", http.StatusUnauthorized, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Login successful", gin.H{"token": token, "candidate": candidate})
}

// Me godoc
// @Summary Get current candidate
// @Tags candidate
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/candidate/me [get]
func (c *CandidateController) Me(ctx *gin.Context) {
	candidate, exists := ctx.Get("currentCandidate")
	if !exists {
		utils.ErrorResponse(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	utils.SuccessResponse(ctx, "Candidate loaded successfully", candidate)
}

// List godoc
// @Summary List candidates
// @Tags admin-candidates
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/candidates [get]
func (c *CandidateController) List(ctx *gin.Context) {
	page, perPage := utils.GetPaginationParams(ctx)
	filters := map[string]string{
		"search": strings.TrimSpace(ctx.Query("search")),
		"status": strings.TrimSpace(ctx.Query("status")),
	}
	startedAt := time.Now()
	result, err := c.service.List(filters, page, perPage)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to load candidates", http.StatusInternalServerError, err.Error())
		return
	}
	utils.APIResponse(ctx, "Candidates loaded successfully", http.StatusOK, result.Items, nil, utils.BuildMeta(result.Pagination, time.Since(startedAt).Seconds()))
}

// Get godoc
// @Summary Get candidate
// @Tags admin-candidates
// @Produce json
// @Security BearerAuth
// @Param id path string true "Candidate ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/candidates/{id} [get]
func (c *CandidateController) Get(ctx *gin.Context) {
	candidate, err := c.service.Get(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Candidate not found", http.StatusNotFound, nil)
		return
	}
	candidate.PasswordHash = ""
	utils.SuccessResponse(ctx, "Candidate loaded successfully", candidate)
}

// Create godoc
// @Summary Create candidate
// @Tags admin-candidates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body dto.CreateCandidateRequest true "Candidate data"
// @Success 201 {object} utils.Response
// @Router /v1/admin/candidates [post]
func (c *CandidateController) Create(ctx *gin.Context) {
	var request dto.CreateCandidateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}
	candidate, err := c.service.Create(request)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to create candidate", http.StatusBadRequest, err.Error())
		return
	}
	utils.CreatedResponse(ctx, "Candidate created successfully", candidate)
}

// Update godoc
// @Summary Update candidate
// @Tags admin-candidates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Candidate ID"
// @Param input body dto.UpdateCandidateRequest true "Candidate data"
// @Success 200 {object} utils.Response
// @Router /v1/admin/candidates/{id} [put]
func (c *CandidateController) Update(ctx *gin.Context) {
	var request dto.UpdateCandidateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}
	candidate, err := c.service.Update(ctx.Param("id"), request)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to update candidate", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Candidate updated successfully", candidate)
}

// Archive godoc
// @Summary Archive candidate
// @Tags admin-candidates
// @Produce json
// @Security BearerAuth
// @Param id path string true "Candidate ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/candidates/{id} [delete]
func (c *CandidateController) Archive(ctx *gin.Context) {
	if err := c.service.Archive(ctx.Param("id")); err != nil {
		utils.ErrorResponse(ctx, "Failed to archive candidate", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Candidate archived successfully", nil)
}
