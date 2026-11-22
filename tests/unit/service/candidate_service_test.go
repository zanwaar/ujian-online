package service_test

import (
	"testing"

	"golang-backend/dto"
	"golang-backend/entity"
	"golang-backend/repository"
	"golang-backend/service"
	"golang-backend/utils"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type fakeCandidateRepository struct {
	candidates  map[string]*entity.Candidate
	created     []entity.Candidate
	createError error
}

func newFakeCandidateRepository() *fakeCandidateRepository {
	return &fakeCandidateRepository{candidates: make(map[string]*entity.Candidate)}
}

func (f *fakeCandidateRepository) List(map[string]string, int, int) (*utils.PaginationResult, error) {
	return nil, nil
}

func (f *fakeCandidateRepository) FindByID(id string) (*entity.Candidate, error) {
	for _, candidate := range f.candidates {
		if candidate.ID == id {
			return candidate, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeCandidateRepository) FindByExamNumber(examNumber string) (*entity.Candidate, error) {
	if candidate, ok := f.candidates[examNumber]; ok {
		return candidate, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeCandidateRepository) Create(candidate *entity.Candidate) error {
	f.candidates[candidate.ExamNumber] = candidate
	return nil
}

func (f *fakeCandidateRepository) CreateBatch(candidates []entity.Candidate) error {
	if f.createError != nil {
		return f.createError
	}
	f.created = append(f.created, candidates...)
	for index := range candidates {
		if candidates[index].ID == "" {
			candidates[index].ID = ulid.Make().String()
		}
		f.candidates[candidates[index].ExamNumber] = &candidates[index]
	}
	return nil
}

func (f *fakeCandidateRepository) Update(candidate *entity.Candidate) error {
	f.candidates[candidate.ExamNumber] = candidate
	return nil
}

func TestCandidateServiceImportHashesPasswordsAndCreatesBatch(t *testing.T) {
	repo := newFakeCandidateRepository()
	rows := []dto.CandidateImportRow{
		{Row: 2, Data: dto.CreateCandidateRequest{ExamNumber: " UJIAN-001 ", Password: "secret1", Name: "Peserta Satu", Gender: "l", School: "SMA 1", ClassName: "X-A"}},
		{Row: 3, Data: dto.CreateCandidateRequest{ExamNumber: "UJIAN-002", Password: "secret2", Name: "Peserta Dua", Gender: "P", School: "SMA 1", ClassName: "X-A"}},
	}

	result, err := service.NewCandidateService(repo).Import(rows)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if result.Imported != 2 || len(repo.created) != 2 {
		t.Fatalf("imported = %d, created = %d; want 2", result.Imported, len(repo.created))
	}
	if repo.created[0].ExamNumber != "UJIAN-001" || repo.created[0].Gender != "L" {
		t.Fatalf("candidate normalization failed: %+v", repo.created[0])
	}
	if repo.created[0].PasswordHash == "secret1" || utils.CheckPassword("secret1", repo.created[0].PasswordHash) != nil {
		t.Fatalf("password was not hashed correctly")
	}
}

func TestCandidateServiceImportRejectsDuplicateRowsWithoutWriting(t *testing.T) {
	repo := newFakeCandidateRepository()
	rows := []dto.CandidateImportRow{
		{Row: 2, Data: dto.CreateCandidateRequest{ExamNumber: "UJIAN-001", Password: "secret1", Name: "Satu", Gender: "L", School: "SMA 1", ClassName: "X-A"}},
		{Row: 3, Data: dto.CreateCandidateRequest{ExamNumber: " UJIAN-001 ", Password: "secret2", Name: "Dua", Gender: "P", School: "SMA 1", ClassName: "X-A"}},
	}

	result, err := service.NewCandidateService(repo).Import(rows)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if result.Imported != 0 || len(result.Errors) != 1 || len(repo.created) != 0 {
		t.Fatalf("unexpected import result: %+v, created=%d", result, len(repo.created))
	}
	if result.Errors[0].Row != 3 {
		t.Fatalf("error row = %d, want 3", result.Errors[0].Row)
	}
}

var _ repository.CandidateRepository = (*fakeCandidateRepository)(nil)
