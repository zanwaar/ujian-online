package migrations

import (
	"log"

	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) {
	log.Println("Running Migrations...")

	migrateRBAC(db)
	migrateUsers(db)
	migrateExamModels(db)

	log.Println("✓ Migrations completed successfully")
}
