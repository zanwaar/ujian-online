package repository

import (
	"golang-backend/entity"
	"golang-backend/utils"

	"gorm.io/gorm"
)

type QuestionRepository interface {
	List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error)
	FindByID(id string) (*entity.Question, error)
	ExamVersionExists(versionID string) (bool, error)
	CreateWithOptions(question *entity.Question, options []entity.QuestionOption) error
	UpdateWithOptions(question *entity.Question, options []entity.QuestionOption) error
	Delete(question *entity.Question) error
}

type questionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository(db *gorm.DB) QuestionRepository {
	return &questionRepository{db: db}
}

func (r *questionRepository) List(filters map[string]string, page, perPage int) (*utils.PaginationResult, error) {
	var questions []entity.Question
	var total int64
	query := r.db.Model(&entity.Question{}).Preload("Options").Where("status <> ?", "archived")

	if versionID := filters["exam_version_id"]; versionID != "" {
		query = query.Where("exam_version_id = ?", versionID)
	}
	if status := filters["status"]; status != "" {
		query = query.Where("status = ?", status)
	}
	if search := filters["search"]; search != "" {
		like := "%" + search + "%"
		query = query.Where("code ILIKE ? OR content_html ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * perPage
	if err := query.Order("display_order ASC, created_at ASC").Limit(perPage).Offset(offset).Find(&questions).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResult{
		Items:      questions,
		Pagination: utils.CalculatePagination(total, page, perPage),
	}, nil
}

func (r *questionRepository) FindByID(id string) (*entity.Question, error) {
	var question entity.Question
	err := r.db.Preload("Options").Where("id = ?", id).First(&question).Error
	return &question, err
}

func (r *questionRepository) ExamVersionExists(versionID string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.ExamVersion{}).Where("id = ? AND status <> ?", versionID, "archived").Count(&count).Error
	return count > 0, err
}

func (r *questionRepository) CreateWithOptions(question *entity.Question, options []entity.QuestionOption) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(question).Error; err != nil {
			return err
		}
		for index := range options {
			options[index].QuestionID = question.ID
			if err := tx.Create(&options[index]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *questionRepository) UpdateWithOptions(question *entity.Question, options []entity.QuestionOption) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(question).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("question_id = ?", question.ID).Delete(&entity.QuestionOption{}).Error; err != nil {
			return err
		}
		for index := range options {
			options[index].QuestionID = question.ID
			if err := tx.Create(&options[index]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *questionRepository) Delete(question *entity.Question) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("question_id = ?", question.ID).Delete(&entity.QuestionOption{}).Error; err != nil {
			return err
		}
		return tx.Delete(question).Error
	})
}
