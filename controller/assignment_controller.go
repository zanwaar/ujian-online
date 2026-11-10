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

type AssignmentController struct {
	service service.AssignmentService
}

func NewAssignmentController(assignmentService service.AssignmentService) *AssignmentController {
	return &AssignmentController{service: assignmentService}
}

func (c *AssignmentController) List(ctx *gin.Context) {
	page, perPage := utils.GetPaginationParams(ctx)
	filters := map[string]string{
		"candidate_id":    strings.TrimSpace(ctx.Query("candidate_id")),
		"exam_id":         strings.TrimSpace(ctx.Query("exam_id")),
		"exam_session_id": strings.TrimSpace(ctx.Query("exam_session_id")),
		"status":          strings.TrimSpace(ctx.Query("status")),
	}
	startedAt := time.Now()
	result, err := c.service.List(filters, page, perPage)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to load assignments", http.StatusInternalServerError, err.Error())
		return
	}
	utils.APIResponse(ctx, "Assignments loaded successfully", http.StatusOK, result.Items, nil, utils.BuildMeta(result.Pagination, time.Since(startedAt).Seconds()))
}

func (c *AssignmentController) Get(ctx *gin.Context) {
	assignment, err := c.service.Get(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Assignment not found", http.StatusNotFound, nil)
		return
	}
	utils.SuccessResponse(ctx, "Assignment loaded successfully", assignment)
}

func (c *AssignmentController) Create(ctx *gin.Context) {
	var request dto.CreateAssignmentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}
	assignment, err := c.service.Create(request)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to create assignment", http.StatusBadRequest, err.Error())
		return
	}
	utils.CreatedResponse(ctx, "Assignment created successfully", assignment)
}

func (c *AssignmentController) Archive(ctx *gin.Context) {
	if err := c.service.Archive(ctx.Param("id")); err != nil {
		utils.ErrorResponse(ctx, "Failed to archive assignment", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Assignment archived successfully", nil)
}
