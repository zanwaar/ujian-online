package service

import (
	"errors"
	"strings"

	"golang-backend/dto"
	"golang-backend/entity"
	"golang-backend/repository"
	"golang-backend/utils"
)

type AssignmentService interface {
	List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error)
	Get(id string) (*entity.ExamAssignment, error)
	Create(req dto.CreateAssignmentRequest) (*entity.ExamAssignment, error)
	Archive(id string) error
}

type assignmentService struct {
	repo repository.AssignmentRepository
}

func NewAssignmentService(repo repository.AssignmentRepository) AssignmentService {
	return &assignmentService{repo: repo}
}

func (s *assignmentService) List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error) {
	return s.repo.List(filters, page, perPage)
}

func (s *assignmentService) Get(id string) (*entity.ExamAssignment, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("assignment id is required")
	}
	return s.repo.FindByID(id)
}

func (s *assignmentService) Create(req dto.CreateAssignmentRequest) (*entity.ExamAssignment, error) {
	if strings.TrimSpace(req.CandidateID) == "" || strings.TrimSpace(req.ExamID) == "" || strings.TrimSpace(req.ExamSessionID) == "" {
		return nil, errors.New("candidate_id, exam_id, and exam_session_id are required")
	}
	candidateExists, err := s.repo.CandidateExists(req.CandidateID)
	if err != nil {
		return nil, err
	}
	if !candidateExists {
		return nil, errors.New("candidate not found or inactive")
	}
	examExists, err := s.repo.ExamExists(req.ExamID)
	if err != nil {
		return nil, err
	}
	if !examExists {
		return nil, errors.New("exam not found")
	}
	sessionMatches, err := s.repo.SessionBelongsToExam(req.ExamSessionID, req.ExamID)
	if err != nil {
		return nil, err
	}
	if !sessionMatches {
		return nil, errors.New("exam session does not belong to exam")
	}
	exists, err := s.repo.IdentityExists(req.CandidateID, req.ExamID, req.ExamSessionID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("assignment already exists")
	}

	assignment := &entity.ExamAssignment{
		CandidateID:   req.CandidateID,
		ExamID:        req.ExamID,
		ExamSessionID: req.ExamSessionID,
		Status:        "assigned",
	}
	if err := s.repo.Create(assignment); err != nil {
		return nil, err
	}
	return assignment, nil
}

func (s *assignmentService) Archive(id string) error {
	assignment, err := s.Get(id)
	if err != nil {
		return err
	}
	return s.repo.Archive(assignment)
}
