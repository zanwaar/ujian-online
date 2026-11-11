package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddlewareUsesIncomingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, RequestID(c))
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Request-ID", "pilot-request-123")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if response.Header().Get("X-Request-ID") != "pilot-request-123" {
		t.Fatalf("expected incoming request ID to be preserved, got %q", response.Header().Get("X-Request-ID"))
	}
	if response.Body.String() != "pilot-request-123" {
		t.Fatalf("expected request ID in context, got %q", response.Body.String())
	}
}

func TestRequestIDMiddlewareReplacesUnsafeID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Request-ID", "unsafe\nrequest")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	requestID := response.Header().Get("X-Request-ID")
	if requestID == "" || requestID == "unsafe\nrequest" {
		t.Fatalf("expected generated safe request ID, got %q", requestID)
	}
}

func TestCORSMiddlewareAllowsConfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware([]string{"https://exam.example.com"}))
	router.OPTIONS("/", func(c *gin.Context) {})

	request := httptest.NewRequest(http.MethodOptions, "/", nil)
	request.Header.Set("Origin", "https://exam.example.com")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", response.Code)
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "https://exam.example.com" {
		t.Fatalf("expected configured origin, got %q", response.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddlewareRejectsUnconfiguredPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware([]string{"https://exam.example.com"}))
	router.OPTIONS("/", func(c *gin.Context) {})

	request := httptest.NewRequest(http.MethodOptions, "/", nil)
	request.Header.Set("Origin", "https://evil.example.com")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", response.Code)
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected no CORS allow header for rejected origin")
	}
}
