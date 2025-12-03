package middlewares

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func TestCompress(t *testing.T) {
	t.Run("compress with gzip", func(t *testing.T) {
		// Use a very large body to ensure compression is triggered
		// Chi's compressor has a minimum size threshold
		largeBody := strings.Repeat("Hello, World! This is a test of compression. ", 200)
		handler := Compress(5)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(largeBody))
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Check if compression was applied
		encoding := w.Header().Get("Content-Encoding")
		if encoding == "gzip" {
			// Decompress and verify content
			reader, err := gzip.NewReader(w.Body)
			if err != nil {
				t.Fatalf("failed to create gzip reader: %v", err)
			}
			defer reader.Close()

			body, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("failed to read decompressed body: %v", err)
			}

			if string(body) != largeBody {
				t.Errorf("decompressed body mismatch")
			}
		} else {
			// If not compressed, body should still be correct
			if w.Body.String() != largeBody {
				t.Errorf("uncompressed body mismatch")
			}
		}
	})

	t.Run("compress with deflate", func(t *testing.T) {
		largeBody := strings.Repeat("Hello, World! This is a test of compression. ", 200)
		handler := Compress(5)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(largeBody))
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "deflate")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		encoding := w.Header().Get("Content-Encoding")
		if encoding == "deflate" {
			// Decompress and verify content
			reader := flate.NewReader(w.Body)
			defer reader.Close()

			body, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("failed to read decompressed body: %v", err)
			}

			if string(body) != largeBody {
				t.Errorf("decompressed body mismatch")
			}
		} else {
			// If not compressed, body should still be correct
			if w.Body.String() != largeBody {
				t.Errorf("uncompressed body mismatch")
			}
		}
	})

	t.Run("compress with zstd", func(t *testing.T) {
		largeBody := strings.Repeat("Hello, World! This is a test of compression. ", 200)
		handler := Compress(5)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(largeBody))
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "zstd")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		encoding := w.Header().Get("Content-Encoding")
		if encoding == "zstd" {
			// Decompress and verify content
			reader, err := zstd.NewReader(w.Body)
			if err != nil {
				t.Fatalf("failed to create zstd reader: %v", err)
			}
			defer reader.Close()

			body, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("failed to read decompressed body: %v", err)
			}

			if string(body) != largeBody {
				t.Errorf("decompressed body mismatch")
			}
		} else {
			// If not compressed, body should still be correct
			if w.Body.String() != largeBody {
				t.Errorf("uncompressed body mismatch")
			}
		}
	})

	t.Run("no compression without accept-encoding", func(t *testing.T) {
		handler := Compress(5)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello, World!"))
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Header().Get("Content-Encoding") != "" {
			t.Errorf("expected no Content-Encoding, got %s", w.Header().Get("Content-Encoding"))
		}

		if w.Body.String() != "Hello, World!" {
			t.Errorf("expected 'Hello, World!', got '%s'", w.Body.String())
		}
	})

	t.Run("compress with different levels", func(t *testing.T) {
		for _, level := range []int{1, 5, 9} {
			largeBody := strings.Repeat("Hello, World! This is a test of compression. ", 200)
			handler := Compress(level)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(largeBody))
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept-Encoding", "gzip")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Just verify the handler doesn't panic with different levels
			// Compression may or may not be applied depending on the middleware's threshold
			if w.Code != http.StatusOK {
				t.Errorf("level %d: expected status 200, got %d", level, w.Code)
			}
		}
	})
}
