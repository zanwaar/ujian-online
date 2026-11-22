package service

import (
	"errors"
	"strconv"
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
	Import(rows []dto.CandidateImportRow) (*dto.CandidateImportResult, error)
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

func (s *candidateService) Import(rows []dto.CandidateImportRow) (*dto.CandidateImportResult, error) {
	result := &dto.CandidateImportResult{Errors: make([]dto.CandidateImportError, 0)}
	if len(rows) == 0 {
		return nil, errors.New("no candidate rows found")
	}

	seen := make(map[string]int, len(rows))
	normalized := make([]dto.CreateCandidateRequest, 0, len(rows))
	for _, row := range rows {
		request := row.Data
		request.ExamNumber = strings.TrimSpace(request.ExamNumber)
		request.Name = strings.TrimSpace(request.Name)
		request.Gender = strings.ToUpper(strings.TrimSpace(request.Gender))
		request.School = strings.TrimSpace(request.School)
		request.ClassName = strings.TrimSpace(request.ClassName)
		request.Password = strings.TrimSpace(request.Password)

		message := validateCandidateImportRequest(request)
		if message == "" {
			if firstRow, exists := seen[request.ExamNumber]; exists {
				message = "exam_number duplicates row " + strconv.Itoa(firstRow)
			} else {
				seen[request.ExamNumber] = row.Row
			}
		}
		if message == "" {
			existing, err := s.repo.FindByExamNumber(request.ExamNumber)
			if err == nil && existing.ID != "" {
				message = "exam_number already exists"
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		}
		if message != "" {
			result.Errors = append(result.Errors, dto.CandidateImportError{Row: row.Row, Message: message})
		}
		normalized = append(normalized, request)
	}
	if len(result.Errors) > 0 {
		return result, nil
	}

	candidates := make([]entity.Candidate, 0, len(normalized))
	for _, request := range normalized {
		hash, err := utils.HashPassword(request.Password)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, entity.Candidate{
			ExamNumber:   request.ExamNumber,
			PasswordHash: hash,
			Name:         request.Name,
			Gender:       request.Gender,
			School:       request.School,
			ClassName:    request.ClassName,
			Status:       "active",
		})
	}
	if err := s.repo.CreateBatch(candidates); err != nil {
		return nil, err
	}
	result.Imported = len(candidates)
	return result, nil
}

func validateCandidateImportRequest(request dto.CreateCandidateRequest) string {
	switch {
	case request.ExamNumber == "":
		return "exam_number is required"
	case len(request.ExamNumber) > 100:
		return "exam_number must be at most 100 characters"
	case len(request.Password) < 6 || len(request.Password) > 100:
		return "password must be between 6 and 100 characters"
	case request.Name == "":
		return "name is required"
	case len(request.Name) > 100:
		return "name must be at most 100 characters"
	case request.Gender != "L" && request.Gender != "P":
		return "gender must be L or P"
	case request.School == "":
		return "school is required"
	case len(request.School) > 100:
		return "school must be at most 100 characters"
	case request.ClassName == "":
		return "class_name is required"
	case len(request.ClassName) > 50:
		return "class_name must be at most 50 characters"
	default:
		return ""
	}
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
