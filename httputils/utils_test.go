package httputils

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/relychan/goutils"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestWriteResponseJSON(t *testing.T) {
	t.Run("write nil body", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := WriteResponseJSON(w, http.StatusOK, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("write json body", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := map[string]string{"message": "hello"}
		err := WriteResponseJSON(w, http.StatusOK, body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var result map[string]string
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if result["message"] != "hello" {
			t.Errorf("expected message 'hello', got '%s'", result["message"])
		}
	})

	t.Run("write json with special characters", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := map[string]string{"message": "<script>alert('xss')</script>"}
		err := WriteResponseJSON(w, http.StatusOK, body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that HTML is not escaped
		bodyStr := w.Body.String()
		if !strings.Contains(bodyStr, "<script>") {
			t.Error("expected HTML not to be escaped")
		}
	})
}

func TestWriteResponseError(t *testing.T) {
	t.Run("write RFC9457 error", func(t *testing.T) {
		w := httptest.NewRecorder()
		rfcErr := goutils.RFC9457Error{
			Type:   "about:blank",
			Title:  "Bad Request",
			Status: http.StatusBadRequest,
			Detail: "Invalid input",
		}
		err := WriteResponseError(w, rfcErr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("write generic error", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := WriteResponseError(w, errors.New("something went wrong"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestDecodeRequestBody(t *testing.T) {
	t.Run("decode valid json", func(t *testing.T) {
		body := `{"name": "test", "value": 123}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		w := httptest.NewRecorder()
		_, span := noop.NewTracerProvider().Tracer("test").Start(context.Background(), "test")

		type TestInput struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		result, ok := DecodeRequestBody[TestInput](w, req, span)
		if !ok {
			t.Fatal("expected decode to succeed")
		}
		if result.Name != "test" {
			t.Errorf("expected name 'test', got '%s'", result.Name)
		}
		if result.Value != 123 {
			t.Errorf("expected value 123, got %d", result.Value)
		}
	})

	t.Run("decode with nil body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()
		_, span := noop.NewTracerProvider().Tracer("test").Start(context.Background(), "test")

		type TestInput struct {
			Name string `json:"name"`
		}

		result, ok := DecodeRequestBody[TestInput](w, req, span)
		if ok {
			t.Error("expected decode to fail with nil body")
		}
		if result != nil {
			t.Error("expected nil result")
		}
	})

	t.Run("decode with invalid json", func(t *testing.T) {
		body := `{"name": "test", invalid}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		w := httptest.NewRecorder()
		_, span := noop.NewTracerProvider().Tracer("test").Start(context.Background(), "test")

		type TestInput struct {
			Name string `json:"name"`
		}

		result, ok := DecodeRequestBody[TestInput](w, req, span)
		if ok {
			t.Error("expected decode to fail with invalid json")
		}
		if result != nil {
			t.Error("expected nil result")
		}
	})
}

func TestGetURLParamUUID(t *testing.T) {
	t.Run("valid uuid", func(t *testing.T) {
		validUUID := uuid.New()
		router := chi.NewRouter()
		router.Get("/test/{id}", func(w http.ResponseWriter, r *http.Request) {
			result, err := GetURLParamUUID(r, "id")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != validUUID {
				t.Errorf("expected uuid %s, got %s", validUUID, result)
			}
		})

		req := httptest.NewRequest("GET", "/test/"+validUUID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		router := chi.NewRouter()
		router.Get("/test/{id}", func(w http.ResponseWriter, r *http.Request) {
			_, err := GetURLParamUUID(r, "id")
			if err == nil {
				t.Error("expected error for invalid uuid")
			}
		})

		req := httptest.NewRequest("GET", "/test/invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})
}

func TestGetURLParamInt64(t *testing.T) {
	t.Run("valid integer", func(t *testing.T) {
		router := chi.NewRouter()
		router.Get("/test/{id}", func(w http.ResponseWriter, r *http.Request) {
			result, err := GetURLParamInt64(r, "id")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != 12345 {
				t.Errorf("expected 12345, got %d", result)
			}
		})

		req := httptest.NewRequest("GET", "/test/12345", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})

	t.Run("invalid integer", func(t *testing.T) {
		router := chi.NewRouter()
		router.Get("/test/{id}", func(w http.ResponseWriter, r *http.Request) {
			_, err := GetURLParamInt64(r, "id")
			if err == nil {
				t.Error("expected error for invalid integer")
			}
		})

		req := httptest.NewRequest("GET", "/test/not-a-number", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})
}

func TestGetRequestLogger(t *testing.T) {
	t.Run("with x-request-id header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("x-request-id", "test-request-id")

		logger := GetRequestLogger(req)
		if logger == nil {
			t.Error("expected logger to be created")
		}
	})

	t.Run("without x-request-id header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)

		logger := GetRequestLogger(req)
		if logger == nil {
			t.Error("expected logger to be created")
		}
	})
}

func TestSetWriteResponseErrorAttribute(t *testing.T) {
	t.Run("set error attribute", func(t *testing.T) {
		_, span := noop.NewTracerProvider().Tracer("test").Start(context.Background(), "test")
		err := errors.New("test error")

		// This should not panic
		SetWriteResponseErrorAttribute(span, err)
	})
}
