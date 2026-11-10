package repository

import (
	"golang-backend/entity"
	"golang-backend/utils"

	"gorm.io/gorm"
)

type AssignmentRepository interface {
	List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error)
	FindByID(id string) (*entity.ExamAssignment, error)
	CandidateExists(candidateID string) (bool, error)
	ExamExists(examID string) (bool, error)
	SessionBelongsToExam(sessionID, examID string) (bool, error)
	IdentityExists(candidateID, examID, sessionID string) (bool, error)
	Create(assignment *entity.ExamAssignment) error
	Archive(assignment *entity.ExamAssignment) error
}

type assignmentRepository struct {
	db *gorm.DB
}

func NewAssignmentRepository(db *gorm.DB) AssignmentRepository {
	return &assignmentRepository{db: db}
}

func (r *assignmentRepository) List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error) {
	var assignments []entity.ExamAssignment
	var total int64
	query := r.db.Model(&entity.ExamAssignment{}).Where("status <> ?", "archived")
	if candidateID := filters["candidate_id"]; candidateID != "" {
		query = query.Where("candidate_id = ?", candidateID)
	}
	if examID := filters["exam_id"]; examID != "" {
		query = query.Where("exam_id = ?", examID)
	}
	if sessionID := filters["exam_session_id"]; sessionID != "" {
		query = query.Where("exam_session_id = ?", sessionID)
	}
	if status := filters["status"]; status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&assignments).Error; err != nil {
		return nil, err
	}
	return &utils.PaginationResult{
		Items:      assignments,
		Pagination: utils.CalculatePagination(total, page, perPage),
	}, nil
}

func (r *assignmentRepository) FindByID(id string) (*entity.ExamAssignment, error) {
	var assignment entity.ExamAssignment
	err := r.db.Where("id = ?", id).First(&assignment).Error
	return &assignment, err
}

func (r *assignmentRepository) CandidateExists(candidateID string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.Candidate{}).Where("id = ? AND status = ?", candidateID, "active").Count(&count).Error
	return count > 0, err
}

func (r *assignmentRepository) ExamExists(examID string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.Exam{}).Where("id = ? AND status <> ?", examID, "archived").Count(&count).Error
	return count > 0, err
}

func (r *assignmentRepository) SessionBelongsToExam(sessionID, examID string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.ExamSession{}).Where("id = ? AND exam_id = ? AND status <> ?", sessionID, examID, "archived").Count(&count).Error
	return count > 0, err
}

func (r *assignmentRepository) IdentityExists(candidateID, examID, sessionID string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.ExamAssignment{}).Where("candidate_id = ? AND exam_id = ? AND exam_session_id = ? AND status <> ?", candidateID, examID, sessionID, "archived").Count(&count).Error
	return count > 0, err
}

func (r *assignmentRepository) Create(assignment *entity.ExamAssignment) error {
	return r.db.Create(assignment).Error
}

func (r *assignmentRepository) Archive(assignment *entity.ExamAssignment) error {
	assignment.Status = "archived"
	return r.db.Save(assignment).Error
}
