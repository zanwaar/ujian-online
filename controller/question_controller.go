package controller

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang-backend/dto"
	"golang-backend/service"
	"golang-backend/utils"

	"github.com/gin-gonic/gin"
)

type QuestionController struct {
	service service.QuestionService
}

func NewQuestionController(questionService service.QuestionService) *QuestionController {
	return &QuestionController{service: questionService}
}

// List godoc
// @Summary List questions
// @Tags admin-questions
// @Produce json
// @Security BearerAuth
// @Param exam_version_id query string false "Exam version ID"
// @Param search query string false "Search term"
// @Param status query string false "Question status"
// @Success 200 {object} utils.Response
// @Router /v1/admin/questions [get]
func (c *QuestionController) List(ctx *gin.Context) {
	page, perPage := utils.GetPaginationParams(ctx)
	filters := map[string]string{
		"exam_version_id": strings.TrimSpace(ctx.Query("exam_version_id")),
		"search":          strings.TrimSpace(ctx.Query("search")),
		"status":          strings.TrimSpace(ctx.Query("status")),
	}
	startedAt := time.Now()
	result, err := c.service.List(filters, page, perPage)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to load questions", http.StatusInternalServerError, err.Error())
		return
	}
	utils.APIResponse(ctx, "Questions loaded successfully", http.StatusOK, result.Items, nil, utils.BuildMeta(result.Pagination, time.Since(startedAt).Seconds()))
}

// Get godoc
// @Summary Get question
// @Tags admin-questions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/questions/{id} [get]
func (c *QuestionController) Get(ctx *gin.Context) {
	question, err := c.service.Get(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Question not found", http.StatusNotFound, nil)
		return
	}
	utils.SuccessResponse(ctx, "Question loaded successfully", question)
}

// Create godoc
// @Summary Create question
// @Tags admin-questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body dto.CreateQuestionRequest true "Question data"
// @Success 201 {object} utils.Response
// @Router /v1/admin/questions [post]
func (c *QuestionController) Create(ctx *gin.Context) {
	var request dto.CreateQuestionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}
	question, err := c.service.Create(request)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to create question", http.StatusBadRequest, err.Error())
		return
	}
	utils.CreatedResponse(ctx, "Question created successfully", question)
}

// Import godoc
// @Summary Import questions from CSV
// @Tags admin-questions
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "CSV file"
// @Success 200 {object} utils.Response
// @Router /v1/admin/questions/import [post]
func (c *QuestionController) Import(ctx *gin.Context) {
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
	for _, column := range []string{"exam_version_id", "code", "content_html", "weight", "display_order", "option_a", "option_b", "correct_option_key"} {
		if _, exists := columnIndexes[column]; !exists {
			utils.ErrorResponse(ctx, "CSV header is invalid", http.StatusBadRequest, fmt.Sprintf("missing column %q", column))
			return
		}
	}

	valueAt := func(record []string, column string) string {
		index := columnIndexes[column]
		if index >= len(record) {
			return ""
		}
		return record[index]
	}
	rows := make([]dto.QuestionImportRow, 0)
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
		if isEmptyQuestionCSVRecord(record) {
			continue
		}
		weight, parseErr := strconv.ParseFloat(strings.TrimSpace(valueAt(record, "weight")), 64)
		if parseErr != nil {
			utils.ErrorResponse(ctx, "CSV value is invalid", http.StatusBadRequest, fmt.Sprintf("row %d: weight must be numeric", rowNumber))
			return
		}
		displayOrder, parseErr := strconv.Atoi(strings.TrimSpace(valueAt(record, "display_order")))
		if parseErr != nil {
			utils.ErrorResponse(ctx, "CSV value is invalid", http.StatusBadRequest, fmt.Sprintf("row %d: display_order must be numeric", rowNumber))
			return
		}
		options := make([]dto.QuestionOptionRequest, 0, 5)
		for index, optionKey := range []string{"A", "B", "C", "D", "E"} {
			content := valueAt(record, "option_"+strings.ToLower(optionKey))
			if strings.TrimSpace(content) == "" {
				continue
			}
			options = append(options, dto.QuestionOptionRequest{OptionKey: optionKey, ContentHTML: content, IsCorrect: strings.EqualFold(optionKey, valueAt(record, "correct_option_key")), DisplayOrder: index + 1})
		}
		rows = append(rows, dto.QuestionImportRow{
			Row: rowNumber,
			Data: dto.CreateQuestionRequest{
				ExamVersionID: valueAt(record, "exam_version_id"),
				Code:          valueAt(record, "code"),
				ContentHTML:   valueAt(record, "content_html"),
				Weight:        weight,
				DisplayOrder:  displayOrder,
				Options:       options,
			},
		})
	}

	result, err := c.service.Import(rows)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to import questions", http.StatusBadRequest, err.Error())
		return
	}
	if len(result.Errors) > 0 {
		utils.ErrorResponse(ctx, "Question import validation failed", http.StatusUnprocessableEntity, result.Errors)
		return
	}
	utils.SuccessResponse(ctx, "Questions imported successfully", result)
}

func isEmptyQuestionCSVRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

// Update godoc
// @Summary Update question
// @Tags admin-questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question ID"
// @Param input body dto.UpdateQuestionRequest true "Question data"
// @Success 200 {object} utils.Response
// @Router /v1/admin/questions/{id} [put]
func (c *QuestionController) Update(ctx *gin.Context) {
	var request dto.UpdateQuestionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}
	question, err := c.service.Update(ctx.Param("id"), request)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to update question", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Question updated successfully", question)
}

// Archive godoc
// @Summary Archive question
// @Tags admin-questions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/questions/{id} [delete]
func (c *QuestionController) Archive(ctx *gin.Context) {
	if err := c.service.Archive(ctx.Param("id")); err != nil {
		utils.ErrorResponse(ctx, "Failed to archive question", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Question archived successfully", nil)
}
