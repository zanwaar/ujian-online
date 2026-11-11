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

// List godoc
// @Summary List assignments
// @Tags admin-assignments
// @Produce json
// @Security BearerAuth
// @Param candidate_id query string false "Candidate ID"
// @Param exam_id query string false "Exam ID"
// @Param exam_session_id query string false "Exam session ID"
// @Param status query string false "Assignment status"
// @Success 200 {object} utils.Response
// @Router /v1/admin/assignments [get]
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

// Get godoc
// @Summary Get assignment
// @Tags admin-assignments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Assignment ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/assignments/{id} [get]
func (c *AssignmentController) Get(ctx *gin.Context) {
	assignment, err := c.service.Get(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Assignment not found", http.StatusNotFound, nil)
		return
	}
	utils.SuccessResponse(ctx, "Assignment loaded successfully", assignment)
}

// Create godoc
// @Summary Create assignment
// @Tags admin-assignments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body dto.CreateAssignmentRequest true "Assignment data"
// @Success 201 {object} utils.Response
// @Router /v1/admin/assignments [post]
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

// Archive godoc
// @Summary Archive assignment
// @Tags admin-assignments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Assignment ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/assignments/{id} [delete]
func (c *AssignmentController) Archive(ctx *gin.Context) {
	if err := c.service.Archive(ctx.Param("id")); err != nil {
		utils.ErrorResponse(ctx, "Failed to archive assignment", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Assignment archived successfully", nil)
}
