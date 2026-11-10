package migrations

import (
	"log"

	"golang-backend/entity"

	"gorm.io/gorm"
)

func migrateExamModels(db *gorm.DB) {
	err := db.AutoMigrate(
		&entity.Exam{},
		&entity.ExamVersion{},
		&entity.ExamSession{},
		&entity.Question{},
		&entity.QuestionOption{},
		&entity.Candidate{},
		&entity.ExamAssignment{},
		&entity.Attempt{},
		&entity.AttemptQuestion{},
		&entity.AttemptAnswer{},
		&entity.AuditLog{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate exam models: %v", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_one_active_attempt_per_assignment ON attempts (assignment_id) WHERE status = 'in_progress' AND deleted_at IS NULL").Error; err != nil {
		log.Fatalf("Failed to create active attempt constraint: %v", err)
	}

	log.Println("✓ Exam models migrated successfully")
}
