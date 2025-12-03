package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMaxBodySize(t *testing.T) {
	t.Run("body within limit", func(t *testing.T) {
		handler := MaxBodySize(1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read body: %v", err)
			}
			if string(body) != "Hello" {
				t.Errorf("expected 'Hello', got '%s'", string(body))
			}
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("Hello")))
		req.ContentLength = 5
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("body exceeds limit by content-length", func(t *testing.T) {
		handler := MaxBodySize(1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Create a body larger than 1KB
		largeBody := strings.Repeat("a", 1025)
		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(largeBody)))
		req.ContentLength = int64(len(largeBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("expected status 413, got %d", w.Code)
		}
	})

	t.Run("zero max body size disables check", func(t *testing.T) {
		handler := MaxBodySize(0)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read body: %v", err)
			}
			if len(body) != 2000 {
				t.Errorf("expected body length 2000, got %d", len(body))
			}
			w.WriteHeader(http.StatusOK)
		}))

		largeBody := strings.Repeat("a", 2000)
		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(largeBody)))
		req.ContentLength = int64(len(largeBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("negative max body size disables check", func(t *testing.T) {
		handler := MaxBodySize(-1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read body: %v", err)
			}
			if len(body) != 2000 {
				t.Errorf("expected body length 2000, got %d", len(body))
			}
			w.WriteHeader(http.StatusOK)
		}))

		largeBody := strings.Repeat("a", 2000)
		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(largeBody)))
		req.ContentLength = int64(len(largeBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("nil body passes through", func(t *testing.T) {
		handler := MaxBodySize(1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("no body passes through", func(t *testing.T) {
		handler := MaxBodySize(1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("exact limit size rejected", func(t *testing.T) {
		handler := MaxBodySize(1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		exactBody := strings.Repeat("a", 1024)
		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(exactBody)))
		req.ContentLength = 1024
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("expected status 413, got %d", w.Code)
		}
	})

	t.Run("one byte under limit allowed", func(t *testing.T) {
		handler := MaxBodySize(1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read body: %v", err)
			}
			if len(body) != 1023 {
				t.Errorf("expected body length 1023, got %d", len(body))
			}
			w.WriteHeader(http.StatusOK)
		}))

		underLimitBody := strings.Repeat("a", 1023)
		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(underLimitBody)))
		req.ContentLength = 1023
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}
