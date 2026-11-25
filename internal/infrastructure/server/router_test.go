package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	apierrors "catalog-service/internal/errors"

	"github.com/gin-gonic/gin"
)

type fakeRoutes struct {
	path       string
	registered bool
}

func (f *fakeRoutes) RegisterRoutes(router *gin.RouterGroup) {
	f.registered = true
	router.GET(f.path, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
}

func TestNewRouterRegistersAllRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	first := &fakeRoutes{path: "/first"}
	second := &fakeRoutes{path: "/second"}

	router := NewRouter(first, second)

	for _, path := range []string{"/api/v1/first", "/api/v1/second"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("GET %s = %d, want %d", path, w.Code, http.StatusOK)
		}
	}
}

type panicRoutes struct{}

func (panicRoutes) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/panic", func(*gin.Context) { panic("boom") })
}

func TestRouterErrorResponsesAreJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	router := NewRouter(&fakeRoutes{path: "/ok"}, panicRoutes{})

	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		code       string
	}{
		{"unknown route", http.MethodGet, "/api/v1/missing", http.StatusNotFound, "ROUTE_NOT_FOUND"},
		{"method not allowed", http.MethodPost, "/api/v1/ok", http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED"},
		{"recovered panic", http.MethodGet, "/api/v1/panic", http.StatusInternalServerError, "INTERNAL_SERVER_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))

			if w.Code != tt.statusCode {
				t.Fatalf("status = %d, want %d", w.Code, tt.statusCode)
			}

			var body apierrors.APIError
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("body is not JSON: %v (%q)", err, w.Body.String())
			}
			if body.Code != tt.code {
				t.Errorf("code = %q, want %q", body.Code, tt.code)
			}
		})
	}
}
