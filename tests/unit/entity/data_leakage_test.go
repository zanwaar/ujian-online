package entity_test

import (
	"encoding/json"
	"strings"
	"testing"

	"golang-backend/entity"
)

func TestSensitiveEntityFieldsAreExcludedFromJSON(t *testing.T) {
	value := struct {
		User      entity.User
		Candidate entity.Candidate
	}{
		User: entity.User{
			Password:         "user-password-hash",
			VerificationCode: "123456",
			ResetToken:       "654321",
			TwoFASecret:      "two-fa-secret",
		},
		Candidate: entity.Candidate{PasswordHash: "candidate-password-hash"},
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	output := string(encoded)
	for _, secret := range []string{"user-password-hash", "123456", "654321", "two-fa-secret", "candidate-password-hash"} {
		if strings.Contains(output, secret) {
			t.Fatalf("serialized entity contains sensitive value %q: %s", secret, output)
		}
	}
}
