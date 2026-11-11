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

type ExamController struct {
	service service.ExamService
}

func NewExamController(examService service.ExamService) *ExamController {
	return &ExamController{service: examService}
}

// List godoc
// @Summary List exams
// @Tags admin-exams
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/exams [get]
func (c *ExamController) List(ctx *gin.Context) {
	page, perPage := utils.GetPaginationParams(ctx)
	filters := map[string]string{
		"search": strings.TrimSpace(ctx.Query("search")),
		"status": strings.TrimSpace(ctx.Query("status")),
	}
	startedAt := time.Now()
	result, err := c.service.List(filters, page, perPage)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to load exams", http.StatusInternalServerError, err.Error())
		return
	}

	utils.APIResponse(ctx, "Exams loaded successfully", http.StatusOK, result.Items, nil, utils.BuildMeta(result.Pagination, time.Since(startedAt).Seconds()))
}

// Get godoc
// @Summary Get exam
// @Tags admin-exams
// @Produce json
// @Security BearerAuth
// @Param id path string true "Exam ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/exams/{id} [get]
func (c *ExamController) Get(ctx *gin.Context) {
	exam, err := c.service.Get(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Exam not found", http.StatusNotFound, nil)
		return
	}
	utils.SuccessResponse(ctx, "Exam loaded successfully", exam)
}

// Create godoc
// @Summary Create exam
// @Tags admin-exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body dto.CreateExamRequest true "Exam data"
// @Success 201 {object} utils.Response
// @Router /v1/admin/exams [post]
func (c *ExamController) Create(ctx *gin.Context) {
	var request dto.CreateExamRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}

	exam, err := c.service.Create(request)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to create exam", http.StatusBadRequest, err.Error())
		return
	}
	utils.CreatedResponse(ctx, "Exam created successfully", exam)
}

// Update godoc
// @Summary Update exam
// @Tags admin-exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Exam ID"
// @Param input body dto.UpdateExamRequest true "Exam data"
// @Success 200 {object} utils.Response
// @Router /v1/admin/exams/{id} [put]
func (c *ExamController) Update(ctx *gin.Context) {
	var request dto.UpdateExamRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}

	exam, err := c.service.Update(ctx.Param("id"), request)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to update exam", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Exam updated successfully", exam)
}

// Archive godoc
// @Summary Archive exam
// @Tags admin-exams
// @Produce json
// @Security BearerAuth
// @Param id path string true "Exam ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/exams/{id} [delete]
func (c *ExamController) Archive(ctx *gin.Context) {
	if err := c.service.Archive(ctx.Param("id")); err != nil {
		utils.ErrorResponse(ctx, "Failed to archive exam", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Exam archived successfully", nil)
}
