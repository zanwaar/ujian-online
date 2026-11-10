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

func (c *CandidateController) Me(ctx *gin.Context) {
	candidate, exists := ctx.Get("currentCandidate")
	if !exists {
		utils.ErrorResponse(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	utils.SuccessResponse(ctx, "Candidate loaded successfully", candidate)
}

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

func (c *CandidateController) Get(ctx *gin.Context) {
	candidate, err := c.service.Get(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Candidate not found", http.StatusNotFound, nil)
		return
	}
	candidate.PasswordHash = ""
	utils.SuccessResponse(ctx, "Candidate loaded successfully", candidate)
}

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

func (c *CandidateController) Archive(ctx *gin.Context) {
	if err := c.service.Archive(ctx.Param("id")); err != nil {
		utils.ErrorResponse(ctx, "Failed to archive candidate", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Candidate archived successfully", nil)
}
