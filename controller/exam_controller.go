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

func (c *ExamController) Get(ctx *gin.Context) {
	exam, err := c.service.Get(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Exam not found", http.StatusNotFound, nil)
		return
	}
	utils.SuccessResponse(ctx, "Exam loaded successfully", exam)
}

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

func (c *ExamController) Archive(ctx *gin.Context) {
	if err := c.service.Archive(ctx.Param("id")); err != nil {
		utils.ErrorResponse(ctx, "Failed to archive exam", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Exam archived successfully", nil)
}
