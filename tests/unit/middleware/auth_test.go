package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang-backend/middleware"
	"golang-backend/utils"

	"github.com/gin-gonic/gin"
)

func runMiddlewareRequest(t *testing.T, handler gin.HandlerFunc, token string) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(handler)
	router.GET("/protected", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	router.ServeHTTP(recorder, request)
	return recorder.Code
}

func TestCandidateMiddlewareRejectsAdminToken(t *testing.T) {
	token, err := utils.GenerateToken("user-1")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if status := runMiddlewareRequest(t, middleware.CandidateAuthMiddleware(), token); status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", status, http.StatusUnauthorized)
	}
}

func TestAdminMiddlewareRejectsCandidateToken(t *testing.T) {
	token, err := utils.GenerateCandidateToken("candidate-1")
	if err != nil {
		t.Fatalf("GenerateCandidateToken() error = %v", err)
	}
	if status := runMiddlewareRequest(t, middleware.AuthMiddleware(), token); status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", status, http.StatusUnauthorized)
	}
}

func TestCandidateMiddlewareRejectsMissingToken(t *testing.T) {
	if status := runMiddlewareRequest(t, middleware.CandidateAuthMiddleware(), ""); status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", status, http.StatusUnauthorized)
	}
}
