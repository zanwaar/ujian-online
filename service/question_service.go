package service

import (
	"errors"
	"fmt"
	"strings"

	"golang-backend/dto"
	"golang-backend/entity"
	"golang-backend/repository"
	"golang-backend/utils"
)

type QuestionService interface {
	List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error)
	Get(id string) (*entity.Question, error)
	Create(req dto.CreateQuestionRequest) (*entity.Question, error)
	Import(rows []dto.QuestionImportRow) (*dto.QuestionImportResult, error)
	Update(id string, req dto.UpdateQuestionRequest) (*entity.Question, error)
	Archive(id string) error
}

type questionService struct {
	repo repository.QuestionRepository
}

func NewQuestionService(repo repository.QuestionRepository) QuestionService {
	return &questionService{repo: repo}
}

func validateQuestionOptions(input []dto.QuestionOptionRequest) ([]entity.QuestionOption, error) {
	if len(input) < 2 || len(input) > 5 {
		return nil, errors.New("a question must have between 2 and 5 options")
	}

	options := make([]entity.QuestionOption, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	correctCount := 0
	for index, item := range input {
		key := strings.ToUpper(strings.TrimSpace(item.OptionKey))
		contentHTML := utils.SanitizeHTML(item.ContentHTML)
		if key == "" || strings.TrimSpace(contentHTML) == "" {
			return nil, fmt.Errorf("option %d requires option_key and content_html", index+1)
		}
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("duplicate option_key: %s", key)
		}
		seen[key] = struct{}{}
		if item.IsCorrect {
			correctCount++
		}
		options = append(options, entity.QuestionOption{
			OptionKey:    key,
			ContentHTML:  contentHTML,
			IsCorrect:    item.IsCorrect,
			DisplayOrder: item.DisplayOrder,
		})
	}
	if correctCount != 1 {
		return nil, errors.New("a question must have exactly one correct option")
	}
	return options, nil
}

func (s *questionService) List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error) {
	return s.repo.List(filters, page, perPage)
}

func (s *questionService) Get(id string) (*entity.Question, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("question id is required")
	}
	return s.repo.FindByID(id)
}

func (s *questionService) Create(req dto.CreateQuestionRequest) (*entity.Question, error) {
	contentHTML := utils.SanitizeHTML(req.ContentHTML)
	if strings.TrimSpace(req.ExamVersionID) == "" || strings.TrimSpace(contentHTML) == "" {
		return nil, errors.New("exam_version_id and content_html are required")
	}
	exists, err := s.repo.ExamVersionExists(req.ExamVersionID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("exam version not found")
	}
	options, err := validateQuestionOptions(req.Options)
	if err != nil {
		return nil, err
	}

	question := &entity.Question{
		ExamVersionID: req.ExamVersionID,
		Code:          strings.TrimSpace(req.Code),
		ContentHTML:   contentHTML,
		Weight:        req.Weight,
		DisplayOrder:  req.DisplayOrder,
		Status:        "active",
	}
	if err := s.repo.CreateWithOptions(question, options); err != nil {
		return nil, err
	}
	question.Options = options
	return question, nil
}

func (s *questionService) Import(rows []dto.QuestionImportRow) (*dto.QuestionImportResult, error) {
	result := &dto.QuestionImportResult{Errors: make([]dto.QuestionImportError, 0)}
	if len(rows) == 0 {
		return nil, errors.New("no question rows found")
	}

	questions := make([]entity.Question, 0, len(rows))
	for _, row := range rows {
		request := row.Data
		request.ExamVersionID = strings.TrimSpace(request.ExamVersionID)
		request.Code = strings.TrimSpace(request.Code)
		request.ContentHTML = strings.TrimSpace(request.ContentHTML)
		message := ""
		if request.ExamVersionID == "" {
			message = "exam_version_id is required"
		} else if request.Weight <= 0 {
			message = "weight must be greater than zero"
		} else if request.DisplayOrder <= 0 {
			message = "display_order must be greater than zero"
		} else if strings.TrimSpace(utils.SanitizeHTML(request.ContentHTML)) == "" {
			message = "content_html is required"
		} else {
			exists, err := s.repo.ExamVersionExists(request.ExamVersionID)
			if err != nil {
				return nil, err
			}
			if !exists {
				message = "exam version not found"
			}
		}

		options, optionsErr := validateQuestionOptions(request.Options)
		if message == "" && optionsErr != nil {
			message = optionsErr.Error()
		}
		if message != "" {
			result.Errors = append(result.Errors, dto.QuestionImportError{Row: row.Row, Message: message})
			continue
		}
		questions = append(questions, entity.Question{
			ExamVersionID: request.ExamVersionID,
			Code:          request.Code,
			ContentHTML:   utils.SanitizeHTML(request.ContentHTML),
			Weight:        request.Weight,
			DisplayOrder:  request.DisplayOrder,
			Status:        "active",
			Options:       options,
		})
	}
	if len(result.Errors) > 0 {
		return result, nil
	}
	if err := s.repo.CreateBatchWithOptions(questions); err != nil {
		return nil, err
	}
	result.Imported = len(questions)
	return result, nil
}

func (s *questionService) Update(id string, req dto.UpdateQuestionRequest) (*entity.Question, error) {
	question, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	options, err := validateQuestionOptions(req.Options)
	if err != nil {
		return nil, err
	}
	contentHTML := utils.SanitizeHTML(req.ContentHTML)
	if strings.TrimSpace(contentHTML) == "" {
		return nil, errors.New("content_html must contain visible text")
	}
	question.Code = strings.TrimSpace(req.Code)
	question.ContentHTML = contentHTML
	question.Weight = req.Weight
	question.DisplayOrder = req.DisplayOrder
	if req.Status != "" {
		question.Status = req.Status
	}
	if err := s.repo.UpdateWithOptions(question, options); err != nil {
		return nil, err
	}
	question.Options = options
	return question, nil
}

func (s *questionService) Archive(id string) error {
	question, err := s.Get(id)
	if err != nil {
		return err
	}
	question.Status = "archived"
	return s.repo.UpdateWithOptions(question, question.Options)
}
