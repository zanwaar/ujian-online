package repository

import (
	"golang-backend/entity"
	"golang-backend/utils"

	"gorm.io/gorm"
)

type CandidateRepository interface {
	List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error)
	FindByID(id string) (*entity.Candidate, error)
	FindByExamNumber(examNumber string) (*entity.Candidate, error)
	Create(candidate *entity.Candidate) error
	Update(candidate *entity.Candidate) error
}

type candidateRepository struct {
	db *gorm.DB
}

func NewCandidateRepository(db *gorm.DB) CandidateRepository {
	return &candidateRepository{db: db}
}

func (r *candidateRepository) List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error) {
	var candidates []entity.Candidate
	var total int64
	query := r.db.Model(&entity.Candidate{}).Where("status <> ?", "archived").Select("id, exam_number, name, gender, school, class_name, status, created_at, updated_at")

	if status := filters["status"]; status != "" {
		query = query.Where("status = ?", status)
	}
	if search := filters["search"]; search != "" {
		like := "%" + search + "%"
		query = query.Where("exam_number ILIKE ? OR name ILIKE ? OR school ILIKE ? OR class_name ILIKE ?", like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&candidates).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResult{
		Items:      candidates,
		Pagination: utils.CalculatePagination(total, page, perPage),
	}, nil
}

func (r *candidateRepository) FindByID(id string) (*entity.Candidate, error) {
	var candidate entity.Candidate
	err := r.db.Where("id = ?", id).First(&candidate).Error
	return &candidate, err
}

func (r *candidateRepository) FindByExamNumber(examNumber string) (*entity.Candidate, error) {
	var candidate entity.Candidate
	err := r.db.Where("exam_number = ?", examNumber).First(&candidate).Error
	return &candidate, err
}

func (r *candidateRepository) Create(candidate *entity.Candidate) error {
	return r.db.Create(candidate).Error
}

func (r *candidateRepository) Update(candidate *entity.Candidate) error {
	return r.db.Save(candidate).Error
}
