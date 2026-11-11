package service_test

import (
	"testing"

	"golang-backend/entity"
	"golang-backend/service"
)

func TestAttemptServiceGetDoesNotExposeResultScore(t *testing.T) {
	repo := newFakeAttemptRepository()
	repo.findAttempt = &entity.Attempt{
		Base:         entity.Base{ID: "attempt-1"},
		Status:       "submitted",
		CorrectCount: 9,
		Score:        90,
		Questions: []entity.AttemptQuestion{{
			Base:                entity.Base{ID: "attempt-question-1"},
			SequenceNumber:      1,
			SnapshotContentJSON: `{"code":"Q1","content_html":"Pertanyaan","correct_option_key":"B","options":[{"option_key":"A","content_html":"Salah"},{"option_key":"B","content_html":"Benar"}]}`,
		}},
	}

	view, err := service.NewAttemptService(repo).Get("candidate-1", "attempt-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if view.CorrectCount != 0 || view.Score != 0 {
		t.Fatalf("candidate attempt view exposes result data: correct_count=%d score=%v", view.CorrectCount, view.Score)
	}
}
