package service

import (
	"errors"
	"strings"

	"golang-backend/dto"
	"golang-backend/entity"
	"golang-backend/repository"
	"golang-backend/utils"
)

type ExamService interface {
	List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error)
	Get(id string) (*entity.Exam, error)
	Create(req dto.CreateExamRequest) (*entity.Exam, error)
	Update(id string, req dto.UpdateExamRequest) (*entity.Exam, error)
	Archive(id string) error
}

type examService struct {
	repo repository.ExamRepository
}

func NewExamService(repo repository.ExamRepository) ExamService {
	return &examService{repo: repo}
}

func (s *examService) List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error) {
	return s.repo.List(filters, page, perPage)
}

func (s *examService) Get(id string) (*entity.Exam, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("exam id is required")
	}
	return s.repo.FindByID(id)
}

func (s *examService) Create(req dto.CreateExamRequest) (*entity.Exam, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Subject) == "" {
		return nil, errors.New("code, name, and subject are required")
	}

	exam := &entity.Exam{
		Code:            code,
		Name:            strings.TrimSpace(req.Name),
		Subject:         strings.TrimSpace(req.Subject),
		DescriptionHTML: req.DescriptionHTML,
		DurationMinutes: req.DurationMinutes,
		ShowResult:      req.ShowResult,
		Status:          "draft",
	}
	version := &entity.ExamVersion{
		VersionNumber:    1,
		ShuffleQuestions: req.ShuffleQuestions,
		ShuffleOptions:   req.ShuffleOptions,
		Status:           "draft",
	}

	if err := s.repo.CreateWithVersion(exam, version); err != nil {
		return nil, err
	}
	return exam, nil
}

func (s *examService) Update(id string, req dto.UpdateExamRequest) (*entity.Exam, error) {
	exam, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Subject) == "" {
		return nil, errors.New("name and subject are required")
	}

	exam.Name = strings.TrimSpace(req.Name)
	exam.Subject = strings.TrimSpace(req.Subject)
	exam.DescriptionHTML = req.DescriptionHTML
	exam.DurationMinutes = req.DurationMinutes
	exam.ShowResult = req.ShowResult
	if req.Status != "" {
		exam.Status = req.Status
	}

	version, err := s.repo.FindCurrentVersion(exam.ID)
	if err != nil {
		return nil, err
	}
	version.ShuffleQuestions = req.ShuffleQuestions
	version.ShuffleOptions = req.ShuffleOptions

	if err := s.repo.Update(exam); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateVersion(version); err != nil {
		return nil, err
	}
	return exam, nil
}

func (s *examService) Archive(id string) error {
	exam, err := s.Get(id)
	if err != nil {
		return err
	}
	exam.Status = "archived"
	return s.repo.Update(exam)
}
