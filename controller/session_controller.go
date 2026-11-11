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

type SessionController struct {
	service service.SessionService
}

func NewSessionController(sessionService service.SessionService) *SessionController {
	return &SessionController{service: sessionService}
}

// List godoc
// @Summary List exam sessions
// @Tags admin-sessions
// @Produce json
// @Security BearerAuth
// @Param exam_id query string false "Exam ID"
// @Param search query string false "Search term"
// @Param status query string false "Session status"
// @Success 200 {object} utils.Response
// @Router /v1/admin/sessions [get]
func (c *SessionController) List(ctx *gin.Context) {
	page, perPage := utils.GetPaginationParams(ctx)
	filters := map[string]string{
		"exam_id": strings.TrimSpace(ctx.Query("exam_id")),
		"search":  strings.TrimSpace(ctx.Query("search")),
		"status":  strings.TrimSpace(ctx.Query("status")),
	}
	startedAt := time.Now()
	result, err := c.service.List(filters, page, perPage)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to load sessions", http.StatusInternalServerError, err.Error())
		return
	}
	utils.APIResponse(ctx, "Sessions loaded successfully", http.StatusOK, result.Items, nil, utils.BuildMeta(result.Pagination, time.Since(startedAt).Seconds()))
}

// Get godoc
// @Summary Get exam session
// @Tags admin-sessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/sessions/{id} [get]
func (c *SessionController) Get(ctx *gin.Context) {
	session, err := c.service.Get(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Session not found", http.StatusNotFound, nil)
		return
	}
	utils.SuccessResponse(ctx, "Session loaded successfully", session)
}

// Create godoc
// @Summary Create exam session
// @Tags admin-sessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body dto.CreateExamSessionRequest true "Session data"
// @Success 201 {object} utils.Response
// @Router /v1/admin/sessions [post]
func (c *SessionController) Create(ctx *gin.Context) {
	var request dto.CreateExamSessionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}
	session, err := c.service.Create(request)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to create session", http.StatusBadRequest, err.Error())
		return
	}
	utils.CreatedResponse(ctx, "Session created successfully", session)
}

// Update godoc
// @Summary Update exam session
// @Tags admin-sessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param input body dto.UpdateExamSessionRequest true "Session data"
// @Success 200 {object} utils.Response
// @Router /v1/admin/sessions/{id} [put]
func (c *SessionController) Update(ctx *gin.Context) {
	var request dto.UpdateExamSessionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}
	session, err := c.service.Update(ctx.Param("id"), request)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to update session", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Session updated successfully", session)
}

// Archive godoc
// @Summary Archive exam session
// @Tags admin-sessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/sessions/{id} [delete]
func (c *SessionController) Archive(ctx *gin.Context) {
	if err := c.service.Archive(ctx.Param("id")); err != nil {
		utils.ErrorResponse(ctx, "Failed to archive session", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Session archived successfully", nil)
}
