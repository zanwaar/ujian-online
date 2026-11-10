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

func (c *SessionController) Get(ctx *gin.Context) {
	session, err := c.service.Get(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Session not found", http.StatusNotFound, nil)
		return
	}
	utils.SuccessResponse(ctx, "Session loaded successfully", session)
}

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

func (c *SessionController) Archive(ctx *gin.Context) {
	if err := c.service.Archive(ctx.Param("id")); err != nil {
		utils.ErrorResponse(ctx, "Failed to archive session", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Session archived successfully", nil)
}
