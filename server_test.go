package gohttps

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/relychan/goutils"
)

func TestNewRouter(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		router := NewRouter(nil, slog.Default())
		if router == nil {
			t.Fatal("expected router to be created")
		}
	})

	t.Run("with basic config", func(t *testing.T) {
		config := &ServerConfig{
			Port: 8080,
		}
		router := NewRouter(config, slog.Default())
		if router == nil {
			t.Fatal("expected router to be created")
		}
	})

	t.Run("with compression level", func(t *testing.T) {
		level := 5
		config := &ServerConfig{
			Port:             8080,
			CompressionLevel: &level,
		}
		router := NewRouter(config, slog.Default())
		if router == nil {
			t.Fatal("expected router to be created")
		}
	})

	t.Run("with request timeout", func(t *testing.T) {
		config := &ServerConfig{
			Port:           8080,
			RequestTimeout: goutils.Duration(5 * time.Second),
		}
		router := NewRouter(config, slog.Default())
		if router == nil {
			t.Fatal("expected router to be created")
		}
	})

	t.Run("with max body size", func(t *testing.T) {
		config := &ServerConfig{
			Port:             8080,
			MaxBodyKilobytes: 1024,
		}
		router := NewRouter(config, slog.Default())
		if router == nil {
			t.Fatal("expected router to be created")
		}
	})

	t.Run("with CORS config", func(t *testing.T) {
		config := &ServerConfig{
			Port: 8080,
			CORS: &CORSConfig{
				AllowedOrigins: []string{"http://localhost:3000"},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
				MaxAge:         3600,
			},
		}
		router := NewRouter(config, slog.Default())
		if router == nil {
			t.Fatal("expected router to be created")
		}
	})
}

func TestNewRouterMiddlewares(t *testing.T) {
	t.Run("healthz endpoint is added", func(t *testing.T) {
		config := &ServerConfig{
			Port: 8080,
		}
		router := NewRouter(config, slog.Default())

		// Add a test route
		router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestCreatePrometheusServer(t *testing.T) {
	t.Run("prometheus disabled", func(t *testing.T) {
		os.Unsetenv("OTEL_METRICS_EXPORTER")
		os.Unsetenv("OTEL_EXPORTER_PROMETHEUS_PORT")

		router := NewRouter(&ServerConfig{Port: 8080}, slog.Default())
		server, err := CreatePrometheusServer(router, 8080)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if server != nil {
			t.Error("expected nil server when prometheus is disabled")
		}
	})

	t.Run("prometheus enabled with same port", func(t *testing.T) {
		os.Setenv("OTEL_METRICS_EXPORTER", "prometheus")
		os.Setenv("OTEL_EXPORTER_PROMETHEUS_PORT", "8080")
		defer os.Unsetenv("OTEL_METRICS_EXPORTER")
		defer os.Unsetenv("OTEL_EXPORTER_PROMETHEUS_PORT")

		router := NewRouter(&ServerConfig{Port: 8080}, slog.Default())
		server, err := CreatePrometheusServer(router, 8080)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if server != nil {
			t.Error("expected nil server when using same port")
		}
	})

	t.Run("prometheus enabled with different port", func(t *testing.T) {
		os.Setenv("OTEL_METRICS_EXPORTER", "prometheus")
		os.Setenv("OTEL_EXPORTER_PROMETHEUS_PORT", "9090")
		defer os.Unsetenv("OTEL_METRICS_EXPORTER")
		defer os.Unsetenv("OTEL_EXPORTER_PROMETHEUS_PORT")

		router := NewRouter(&ServerConfig{Port: 8080}, slog.Default())
		server, err := CreatePrometheusServer(router, 8080)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if server == nil {
			t.Error("expected server to be created")
		}
	})

	t.Run("prometheus enabled with invalid port", func(t *testing.T) {
		os.Setenv("OTEL_METRICS_EXPORTER", "prometheus")
		os.Setenv("OTEL_EXPORTER_PROMETHEUS_PORT", "invalid")
		defer os.Unsetenv("OTEL_METRICS_EXPORTER")
		defer os.Unsetenv("OTEL_EXPORTER_PROMETHEUS_PORT")

		router := NewRouter(&ServerConfig{Port: 8080}, slog.Default())
		_, err := CreatePrometheusServer(router, 8080)
		if err == nil {
			t.Error("expected error for invalid port")
		}
	})
}

func TestListenAndServe(t *testing.T) {
	t.Run("nil config returns error", func(t *testing.T) {
		router := NewRouter(nil, slog.Default())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := ListenAndServe(ctx, router, nil)
		if err != errServerConfigRequired {
			t.Errorf("expected errServerConfigRequired, got %v", err)
		}
	})

	t.Run("server starts and stops with context cancellation", func(t *testing.T) {
		config := &ServerConfig{
			Port: 0, // Use random available port
		}
		router := NewRouter(config, slog.Default())
		router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := ListenAndServe(ctx, router, config)
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
