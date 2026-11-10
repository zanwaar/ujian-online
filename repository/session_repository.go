package repository

import (
	"golang-backend/entity"
	"golang-backend/utils"

	"gorm.io/gorm"
)

type SessionRepository interface {
	List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error)
	FindByID(id string) (*entity.ExamSession, error)
	ExamExists(examID string) (bool, error)
	Create(session *entity.ExamSession) error
	Update(session *entity.ExamSession) error
	Delete(session *entity.ExamSession) error
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error) {
	var sessions []entity.ExamSession
	var total int64
	query := r.db.Model(&entity.ExamSession{}).Where("status <> ?", "archived")

	if examID := filters["exam_id"]; examID != "" {
		query = query.Where("exam_id = ?", examID)
	}
	if status := filters["status"]; status != "" {
		query = query.Where("status = ?", status)
	}
	if search := filters["search"]; search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * perPage
	if err := query.Order("starts_at ASC").Limit(perPage).Offset(offset).Find(&sessions).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResult{
		Items:      sessions,
		Pagination: utils.CalculatePagination(total, page, perPage),
	}, nil
}

func (r *sessionRepository) FindByID(id string) (*entity.ExamSession, error) {
	var session entity.ExamSession
	err := r.db.Where("id = ?", id).First(&session).Error
	return &session, err
}

func (r *sessionRepository) ExamExists(examID string) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.Exam{}).Where("id = ? AND status <> ?", examID, "archived").Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *sessionRepository) Create(session *entity.ExamSession) error {
	return r.db.Create(session).Error
}

func (r *sessionRepository) Update(session *entity.ExamSession) error {
	return r.db.Save(session).Error
}

func (r *sessionRepository) Delete(session *entity.ExamSession) error {
	return r.db.Delete(session).Error
}
