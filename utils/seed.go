package utils

import (
	"golang-backend/entity"
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
)

func SeedRolesAndPermissions(db *gorm.DB) {
	// Define permissions
	permissions := []string{
		"manage_users",
		"manage_roles",
		"view_reports",
	}

	var createdPermissions []*entity.Permission
	for _, permName := range permissions {
		perm := &entity.Permission{Name: permName}
		if err := db.FirstOrCreate(perm, entity.Permission{Name: permName}).Error; err != nil {
			log.Printf("Failed to seed permission %s: %v", permName, err)
		} else {
			createdPermissions = append(createdPermissions, perm)
		}
	}

	// Define Roles
	roles := []string{"admin", "user", "manager"}

	for _, roleName := range roles {
		role := &entity.Role{Name: roleName}
		if err := db.FirstOrCreate(role, entity.Role{Name: roleName}).Error; err != nil {
			log.Printf("Failed to seed role %s: %v", roleName, err)
		}

		// Assign permissions to admin
		if roleName == "admin" {
			if err := db.Model(role).Association("Permissions").Replace(createdPermissions); err != nil {
				log.Printf("Failed to replace permissions for role %s: %v", roleName, err)
			}
		}
	}

	log.Println("✓ Roles and Permissions seeded successfully")
}

func SeedUsers(db *gorm.DB) {
	// Fetch Roles
	var adminRole, userRole, managerRole entity.Role
	db.Where("name = ?", "admin").First(&adminRole)
	db.Where("name = ?", "user").First(&userRole)
	db.Where("name = ?", "manager").First(&managerRole)

	// Pre-hash password "password" to speed up seeding
	hashedPassword, _ := HashPassword("password")

	log.Println("🌱 Seeding users...")

	// 1. Super Admin
	superAdmin := entity.User{
		Name:       "Super Admin",
		Email:      "superadmin@example.com",
		IsVerified: true,
	}
	// Use FirstOrCreate to ensure existence.
	// Note: We need to assign password and roles if creating new.
	// GORM's FirstOrCreate is tricky with related data (Roles).
	// Let's check by email first.
	var existingSuperAdmin entity.User
	if err := db.Where("email = ?", superAdmin.Email).First(&existingSuperAdmin).Error; err != nil {
		// Not found, create
		superAdmin.Password = hashedPassword
		superAdmin.Roles = []*entity.Role{&adminRole, &managerRole}
		if err := db.Create(&superAdmin).Error; err != nil {
			log.Printf("Failed to seed Super Admin: %v", err)
		} else {
			log.Println("✓ Super Admin seeded")
		}
	}

	// 2. Admin
	admin := entity.User{
		Name:       "Admin",
		Email:      "admin@example.com",
		IsVerified: true,
	}
	var existingAdmin entity.User
	if err := db.Where("email = ?", admin.Email).First(&existingAdmin).Error; err != nil {
		admin.Password = hashedPassword
		admin.Roles = []*entity.Role{&adminRole}
		if err := db.Create(&admin).Error; err != nil {
			log.Printf("Failed to seed Admin: %v", err)
		} else {
			log.Println("✓ Admin seeded")
		}
	}

	// 3. Regular Users (Only if count is low)
	var count int64
	db.Model(&entity.User{}).Count(&count)
	if count > 10 {
		return
	}

	for i := 1; i <= 50; i++ {
		email := "user" + strconv.Itoa(i) + "@example.com"
		user := entity.User{
			Name:       "User " + strconv.Itoa(i),
			Email:      email,
			Password:   hashedPassword,
			IsVerified: true,
			Roles:      []*entity.Role{&userRole},
		}

		if err := db.Create(&user).Error; err != nil {
			log.Printf("Failed to seed user %s: %v", email, err)
		}
	}
	log.Println("✓ Users seeded successfully")
}

func SeedDemoExam(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		const examCode = "DEMO-UTS-001"
		const candidateNumber = "DEMO-001"

		var exam entity.Exam
		if err := tx.Where("code = ?", examCode).First(&exam).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			exam = entity.Exam{
				Code:            examCode,
				Name:            "Ujian Demo Go React",
				Subject:         "Pemrograman Dasar",
				DescriptionHTML: "Ujian demo untuk memverifikasi alur kandidat dari login sampai hasil.",
				DurationMinutes: 60,
				ShowResult:      true,
				Status:          "published",
			}
			if err := tx.Create(&exam).Error; err != nil {
				return err
			}
		}

		var version entity.ExamVersion
		if err := tx.Where("exam_id = ? AND version_number = ?", exam.ID, 1).First(&version).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			now := time.Now().UTC()
			version = entity.ExamVersion{
				ExamID:           exam.ID,
				VersionNumber:    1,
				ShuffleQuestions: false,
				ShuffleOptions:   false,
				Status:           "published",
				PublishedAt:      &now,
			}
			if err := tx.Create(&version).Error; err != nil {
				return err
			}
		}
		if exam.CurrentVersionID == nil || *exam.CurrentVersionID != version.ID {
			exam.CurrentVersionID = &version.ID
			if err := tx.Model(&exam).Update("current_version_id", version.ID).Error; err != nil {
				return err
			}
		}

		questions := []struct {
			code    string
			content string
			answer  string
		}{
			{code: "DEMO-Q-01", content: "Bahasa yang digunakan oleh backend target adalah ...", answer: "B"},
			{code: "DEMO-Q-02", content: "HTTP method untuk menyimpan jawaban secara idempotent adalah ...", answer: "C"},
		}
		for index, item := range questions {
			var question entity.Question
			if err := tx.Where("exam_version_id = ? AND code = ?", version.ID, item.code).First(&question).Error; err != nil {
				if err != gorm.ErrRecordNotFound {
					return err
				}
				question = entity.Question{
					ExamVersionID: version.ID,
					Code:          item.code,
					ContentHTML:   item.content,
					Weight:        1,
					DisplayOrder:  index + 1,
					Status:        "active",
				}
				if err := tx.Create(&question).Error; err != nil {
					return err
				}
			}
			options := []struct {
				key     string
				content string
			}{
				{key: "A", content: "PHP"},
				{key: "B", content: "Go"},
				{key: "C", content: "Ruby"},
				{key: "D", content: "Python"},
			}
			if item.code == "DEMO-Q-02" {
				options[0].content = "GET"
				options[1].content = "POST"
				options[2].content = "PUT"
				options[3].content = "DELETE"
			}
			for optionIndex, option := range options {
				var questionOption entity.QuestionOption
				if err := tx.Where("question_id = ? AND option_key = ?", question.ID, option.key).First(&questionOption).Error; err != nil {
					if err != gorm.ErrRecordNotFound {
						return err
					}
					if err := tx.Create(&entity.QuestionOption{
						QuestionID:   question.ID,
						OptionKey:    option.key,
						ContentHTML:  option.content,
						IsCorrect:    option.key == item.answer,
						DisplayOrder: optionIndex + 1,
					}).Error; err != nil {
						return err
					}
				}
			}
		}

		var session entity.ExamSession
		if err := tx.Where("exam_id = ? AND name = ?", exam.ID, "Demo Session").First(&session).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			now := time.Now().UTC()
			session = entity.ExamSession{
				ExamID:   exam.ID,
				Name:     "Demo Session",
				StartsAt: now.Add(-5 * time.Minute),
				EndsAt:   now.Add(2 * time.Hour),
				Status:   "open",
			}
			if err := tx.Create(&session).Error; err != nil {
				return err
			}
		}

		var candidate entity.Candidate
		if err := tx.Where("exam_number = ?", candidateNumber).First(&candidate).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			passwordHash, err := HashPassword("password")
			if err != nil {
				return err
			}
			candidate = entity.Candidate{
				ExamNumber:   candidateNumber,
				PasswordHash: passwordHash,
				Name:         "Peserta Demo",
				Gender:       "L",
				School:       "Sekolah Demo",
				ClassName:    "XII IPA 1",
				Status:       "active",
			}
			if err := tx.Create(&candidate).Error; err != nil {
				return err
			}
		}

		var assignment entity.ExamAssignment
		assignmentQuery := tx.Where("candidate_id = ? AND exam_id = ? AND exam_session_id = ?", candidate.ID, exam.ID, session.ID)
		if err := assignmentQuery.First(&assignment).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			assignment = entity.ExamAssignment{
				CandidateID:   candidate.ID,
				ExamID:        exam.ID,
				ExamSessionID: session.ID,
				Status:        "assigned",
			}
			if err := tx.Create(&assignment).Error; err != nil {
				return err
			}
		}

		log.Printf("✓ Demo data ready: exam=%s candidate=%s password=password assignment=%s", examCode, candidateNumber, assignment.ID)
		return nil
	})
}
