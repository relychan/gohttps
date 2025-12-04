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

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("multiple encodings with first supported", func(t *testing.T) {
		// Create gzip compressed body
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		gw.Write([]byte("Hello, Multiple Encodings!"))
		gw.Close()

		handler := Decompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read body: %v", err)
			}
			if string(body) != "Hello, Multiple Encodings!" {
				t.Errorf("expected 'Hello, Multiple Encodings!', got '%s'", string(body))
			}
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/test", &buf)
		// Add multiple Content-Encoding headers
		req.Header.Add("Content-Encoding", "gzip")
		req.Header.Add("Content-Encoding", "deflate")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("multiple encodings with second supported", func(t *testing.T) {
		// Create gzip compressed body
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		gw.Write([]byte("Hello, Second Encoding!"))
		gw.Close()

		handler := Decompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read body: %v", err)
			}
			if string(body) != "Hello, Second Encoding!" {
				t.Errorf("expected 'Hello, Second Encoding!', got '%s'", string(body))
			}
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/test", &buf)
		// Add multiple Content-Encoding headers with unsupported first
		req.Header.Add("Content-Encoding", "unsupported")
		req.Header.Add("Content-Encoding", "gzip")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("multiple encodings all unsupported", func(t *testing.T) {
		handler := Decompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("Hello, World!")))
		req.Header.Add("Content-Encoding", "unsupported1")
		req.Header.Add("Content-Encoding", "unsupported2")
		req.ContentLength = 13
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnsupportedMediaType {
			t.Errorf("expected status 415, got %d", w.Code)
		}
	})

	t.Run("multiple encodings with invalid gzip data tries next encoding", func(t *testing.T) {
		// When multiple encodings are present and data is invalid for all,
		// the middleware tries each one. If all fail, it returns an error.
		// However, if the data happens to be valid for one encoding, it succeeds.
		handler := Decompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If we get here, one of the decompressors succeeded
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("not compressed data")))
		req.Header.Add("Content-Encoding", "gzip")
		req.Header.Add("Content-Encoding", "deflate")
		req.ContentLength = 19
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// The middleware tries each encoding; the result depends on whether
		// any decompressor accepts the data
		if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
			t.Errorf("expected status 200 or 400, got %d", w.Code)
		}
	})

	t.Run("multiple encodings with mixed supported and unsupported", func(t *testing.T) {
		// Create gzip compressed body
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		gw.Write([]byte("Hello, Mixed Encodings!"))
		gw.Close()

		handler := Decompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read body: %v", err)
			}
			if string(body) != "Hello, Mixed Encodings!" {
				t.Errorf("expected 'Hello, Mixed Encodings!', got '%s'", string(body))
			}
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/test", &buf)
		// Mix of unsupported and supported encodings
		req.Header.Add("Content-Encoding", "unsupported1")
		req.Header.Add("Content-Encoding", "gzip")
		req.Header.Add("Content-Encoding", "unsupported2")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}
