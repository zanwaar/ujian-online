package service

import (
	"errors"
	"strings"

	"golang-backend/dto"
	"golang-backend/entity"
	"golang-backend/repository"
	"golang-backend/utils"
)

type SessionService interface {
	List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error)
	Get(id string) (*entity.ExamSession, error)
	Create(req dto.CreateExamSessionRequest) (*entity.ExamSession, error)
	Update(id string, req dto.UpdateExamSessionRequest) (*entity.ExamSession, error)
	Archive(id string) error
}

type sessionService struct {
	repo repository.SessionRepository
}

func NewSessionService(repo repository.SessionRepository) SessionService {
	return &sessionService{repo: repo}
}

func (s *sessionService) List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error) {
	return s.repo.List(filters, page, perPage)
}

func (s *sessionService) Get(id string) (*entity.ExamSession, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("session id is required")
	}
	return s.repo.FindByID(id)
}

func (s *sessionService) Create(req dto.CreateExamSessionRequest) (*entity.ExamSession, error) {
	if strings.TrimSpace(req.ExamID) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("exam_id and name are required")
	}
	if !req.EndsAt.After(req.StartsAt) {
		return nil, errors.New("ends_at must be after starts_at")
	}
	exists, err := s.repo.ExamExists(req.ExamID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("exam not found")
	}

	session := &entity.ExamSession{
		ExamID:   req.ExamID,
		Name:     strings.TrimSpace(req.Name),
		StartsAt: req.StartsAt,
		EndsAt:   req.EndsAt,
		Status:   "scheduled",
	}
	if err := s.repo.Create(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *sessionService) Update(id string, req dto.UpdateExamSessionRequest) (*entity.ExamSession, error) {
	session, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("name is required")
	}
	if !req.EndsAt.After(req.StartsAt) {
		return nil, errors.New("ends_at must be after starts_at")
	}

	session.Name = strings.TrimSpace(req.Name)
	session.StartsAt = req.StartsAt
	session.EndsAt = req.EndsAt
	if req.Status != "" {
		session.Status = req.Status
	}
	if err := s.repo.Update(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *sessionService) Archive(id string) error {
	session, err := s.Get(id)
	if err != nil {
		return err
	}
	session.Status = "archived"
	return s.repo.Update(session)
}
