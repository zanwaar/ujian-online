package controller

import (
	"encoding/csv"
	"fmt"
	"io"
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

// Import godoc
// @Summary Import candidates from CSV
// @Tags admin-candidates
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "CSV file"
// @Success 200 {object} utils.Response
// @Router /v1/admin/candidates/import [post]
func (c *CandidateController) Import(ctx *gin.Context) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 5<<20)
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		utils.ErrorResponse(ctx, "CSV file is required", http.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()

	reader := csv.NewReader(io.LimitReader(file, 5<<20+1))
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		utils.ErrorResponse(ctx, "CSV header is required", http.StatusBadRequest, err.Error())
		return
	}

	columnIndexes := make(map[string]int, len(header))
	for index, column := range header {
		column = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(column, "\ufeff")))
		if column != "" {
			columnIndexes[column] = index
		}
	}
	requiredColumns := []string{"exam_number", "password", "name", "gender", "school", "class_name"}
	for _, column := range requiredColumns {
		if _, exists := columnIndexes[column]; !exists {
			utils.ErrorResponse(ctx, "CSV header is invalid", http.StatusBadRequest, fmt.Sprintf("missing column %q", column))
			return
		}
	}

	rows := make([]dto.CandidateImportRow, 0)
	rowNumber := 1
	for {
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		rowNumber++
		if readErr != nil {
			utils.ErrorResponse(ctx, "CSV could not be read", http.StatusBadRequest, fmt.Sprintf("row %d: %v", rowNumber, readErr))
			return
		}
		if isEmptyCSVRecord(record) {
			continue
		}
		value := func(column string) string {
			index := columnIndexes[column]
			if index >= len(record) {
				return ""
			}
			return record[index]
		}
		rows = append(rows, dto.CandidateImportRow{
			Row: rowNumber,
			Data: dto.CreateCandidateRequest{
				ExamNumber: value("exam_number"),
				Password:   value("password"),
				Name:       value("name"),
				Gender:     value("gender"),
				School:     value("school"),
				ClassName:  value("class_name"),
			},
		})
	}

	result, err := c.service.Import(rows)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to import candidates", http.StatusBadRequest, err.Error())
		return
	}
	if len(result.Errors) > 0 {
		utils.ErrorResponse(ctx, "Candidate import validation failed", http.StatusUnprocessableEntity, result.Errors)
		return
	}
	utils.SuccessResponse(ctx, "Candidates imported successfully", result)
}

func isEmptyCSVRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
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
