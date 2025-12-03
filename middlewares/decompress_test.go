package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecompress(t *testing.T) {
	t.Run("decompress gzip body", func(t *testing.T) {
		// Create gzip compressed body
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		gw.Write([]byte("Hello, World!"))
		gw.Close()

		handler := Decompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read body: %v", err)
			}
			if string(body) != "Hello, World!" {
				t.Errorf("expected 'Hello, World!', got '%s'", string(body))
			}
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/test", &buf)
		req.Header.Set("Content-Encoding", "gzip")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("no decompression without content-encoding", func(t *testing.T) {
		handler := Decompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read body: %v", err)
			}
			if string(body) != "Hello, World!" {
				t.Errorf("expected 'Hello, World!', got '%s'", string(body))
			}
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("Hello, World!")))
		req.ContentLength = 13
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("skip decompression for nil body", func(t *testing.T) {
		handler := Decompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Content-Encoding", "gzip")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("skip decompression for empty content length", func(t *testing.T) {
		handler := Decompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte{}))
		req.Header.Set("Content-Encoding", "gzip")
		req.ContentLength = 0
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("unsupported encoding returns 415", func(t *testing.T) {
		handler := Decompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("Hello, World!")))
		req.Header.Set("Content-Encoding", "unsupported-encoding")
		req.ContentLength = 13
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnsupportedMediaType {
			t.Errorf("expected status 415, got %d", w.Code)
		}
	})

	t.Run("invalid gzip data returns error", func(t *testing.T) {
		handler := Decompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("not gzip data")))
		req.Header.Set("Content-Encoding", "gzip")
		req.ContentLength = 13
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}
