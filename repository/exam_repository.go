package repository

import (
	"golang-backend/entity"
	"golang-backend/utils"

	"gorm.io/gorm"
)

type ExamRepository interface {
	List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error)
	FindByID(id string) (*entity.Exam, error)
	FindCurrentVersion(examID string) (*entity.ExamVersion, error)
	CreateWithVersion(exam *entity.Exam, version *entity.ExamVersion) error
	Update(exam *entity.Exam) error
	UpdateVersion(version *entity.ExamVersion) error
	Delete(exam *entity.Exam) error
}

type examRepository struct {
	db *gorm.DB
}

func NewExamRepository(db *gorm.DB) ExamRepository {
	return &examRepository{db: db}
}

func (r *examRepository) List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error) {
	var exams []entity.Exam
	var total int64
	query := r.db.Model(&entity.Exam{}).Where("status <> ?", "archived")

	if search := filters["search"]; search != "" {
		like := "%" + search + "%"
		query = query.Where("code ILIKE ? OR name ILIKE ? OR subject ILIKE ?", like, like, like)
	}
	if status := filters["status"]; status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * perPage
	if err := query.Preload("Versions", "status <> ?", "archived").Order("created_at DESC").Limit(perPage).Offset(offset).Find(&exams).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResult{
		Items:      exams,
		Pagination: utils.CalculatePagination(total, page, perPage),
	}, nil
}

func (r *examRepository) FindByID(id string) (*entity.Exam, error) {
	var exam entity.Exam
	err := r.db.Preload("Versions", "status <> ?", "archived").Where("id = ?", id).First(&exam).Error
	return &exam, err
}

func (r *examRepository) FindCurrentVersion(examID string) (*entity.ExamVersion, error) {
	var exam entity.Exam
	if err := r.db.Select("current_version_id").Where("id = ?", examID).First(&exam).Error; err != nil {
		return nil, err
	}
	if exam.CurrentVersionID == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var version entity.ExamVersion
	err := r.db.Where("id = ?", *exam.CurrentVersionID).First(&version).Error
	return &version, err
}

func (r *examRepository) CreateWithVersion(exam *entity.Exam, version *entity.ExamVersion) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(exam).Error; err != nil {
			return err
		}
		version.ExamID = exam.ID
		if err := tx.Create(version).Error; err != nil {
			return err
		}
		exam.CurrentVersionID = &version.ID
		return tx.Model(exam).Update("current_version_id", version.ID).Error
	})
}

func (r *examRepository) Update(exam *entity.Exam) error {
	return r.db.Save(exam).Error
}

func (r *examRepository) UpdateVersion(version *entity.ExamVersion) error {
	return r.db.Save(version).Error
}

func (r *examRepository) Delete(exam *entity.Exam) error {
	return r.db.Delete(exam).Error
}
