package service

import (
	"errors"
	"strings"

	"golang-backend/dto"
	"golang-backend/entity"
	"golang-backend/repository"
	"golang-backend/utils"

	"gorm.io/gorm"
)

type CandidateService interface {
	Login(req dto.CandidateLoginRequest) (string, *entity.Candidate, error)
	List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error)
	Get(id string) (*entity.Candidate, error)
	Create(req dto.CreateCandidateRequest) (*entity.Candidate, error)
	Update(id string, req dto.UpdateCandidateRequest) (*entity.Candidate, error)
	Archive(id string) error
}

type candidateService struct {
	repo repository.CandidateRepository
}

func NewCandidateService(repo repository.CandidateRepository) CandidateService {
	return &candidateService{repo: repo}
}

func (s *candidateService) Login(req dto.CandidateLoginRequest) (string, *entity.Candidate, error) {
	candidate, err := s.repo.FindByExamNumber(strings.TrimSpace(req.ExamNumber))
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}
	if candidate.Status != "active" || utils.CheckPassword(req.Password, candidate.PasswordHash) != nil {
		return "", nil, errors.New("invalid credentials")
	}
	token, err := utils.GenerateCandidateToken(candidate.ID)
	if err != nil {
		return "", nil, err
	}
	candidate.PasswordHash = ""
	return token, candidate, nil
}

func (s *candidateService) List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error) {
	return s.repo.List(filters, page, perPage)
}

func (s *candidateService) Get(id string) (*entity.Candidate, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("candidate id is required")
	}
	return s.repo.FindByID(id)
}

func (s *candidateService) Create(req dto.CreateCandidateRequest) (*entity.Candidate, error) {
	examNumber := strings.TrimSpace(req.ExamNumber)
	if examNumber == "" || strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("exam_number and name are required")
	}
	if existing, err := s.repo.FindByExamNumber(examNumber); err == nil && existing.ID != "" {
		return nil, errors.New("exam number already exists")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	candidate := &entity.Candidate{
		ExamNumber:   examNumber,
		PasswordHash: hash,
		Name:         strings.TrimSpace(req.Name),
		Gender:       req.Gender,
		School:       strings.TrimSpace(req.School),
		ClassName:    strings.TrimSpace(req.ClassName),
		Status:       "active",
	}
	if err := s.repo.Create(candidate); err != nil {
		return nil, err
	}
	candidate.PasswordHash = ""
	return candidate, nil
}

func (s *candidateService) Update(id string, req dto.UpdateCandidateRequest) (*entity.Candidate, error) {
	candidate, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	candidate.Name = strings.TrimSpace(req.Name)
	candidate.Gender = req.Gender
	candidate.School = strings.TrimSpace(req.School)
	candidate.ClassName = strings.TrimSpace(req.ClassName)
	if req.Status != "" {
		candidate.Status = req.Status
	}
	if req.Password != "" {
		candidate.PasswordHash, err = utils.HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
	}
	if err := s.repo.Update(candidate); err != nil {
		return nil, err
	}
	candidate.PasswordHash = ""
	return candidate, nil
}

func (s *candidateService) Archive(id string) error {
	candidate, err := s.Get(id)
	if err != nil {
		return err
	}
	candidate.Status = "archived"
	return s.repo.Update(candidate)
}
