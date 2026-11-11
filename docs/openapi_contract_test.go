package docs

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCriticalOpenAPIPaths(t *testing.T) {
	data, err := os.ReadFile("swagger.json")
	if err != nil {
		t.Fatalf("read swagger.json: %v", err)
	}

	var document struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode swagger.json: %v", err)
	}

	requiredOperations := map[string][]string{
		"/v1/auth/candidate/login":                          {"post"},
		"/v1/candidate/me":                                  {"get"},
		"/v1/candidate/exams":                               {"get"},
		"/v1/candidate/exams/{exam_id}/attempts":            {"post"},
		"/v1/candidate/attempts/{id}":                       {"get"},
		"/v1/candidate/attempts/{id}/answers/{question_id}": {"put"},
		"/v1/candidate/attempts/{id}/submit":                {"post"},
		"/v1/candidate/attempts/{id}/result":                {"get"},
		"/v1/admin/exams":                                   {"get", "post"},
		"/v1/admin/questions":                               {"get", "post"},
		"/v1/admin/candidates":                              {"get", "post"},
		"/v1/admin/assignments":                             {"get", "post"},
		"/v1/admin/results":                                 {"get"},
		"/v1/admin/results/{id}":                            {"get"},
	}

	for path, methods := range requiredOperations {
		operations, ok := document.Paths[path]
		if !ok {
			t.Errorf("OpenAPI path %q is missing", path)
			continue
		}
		for _, method := range methods {
			if _, ok := operations[method]; !ok {
				t.Errorf("OpenAPI operation %s %s is missing", method, path)
			}
		}
	}
}
