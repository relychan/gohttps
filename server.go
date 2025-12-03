// Package gohttps includes reusable functions to create HTTP servers in Go.
package gohttps

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter creates a new router with default middlewares.
func NewRouter(envVars ServerConfig, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)

	if envVars.RequestTimeout > 0 {
		router.Use(middleware.Timeout(time.Duration(envVars.RequestTimeout)))
	}

	if envVars.CompressionLevel > 0 {
		router.Use(middleware.Compress(envVars.CompressionLevel))
	}

	if envVars.MaxBodyKilobytes > 0 {
		router.Use(MaxBodySizeMiddleware(envVars.MaxBodyKilobytes))
	}

	if envVars.CORS != nil && len(envVars.CORS.AllowedOrigins) > 0 {
		router.Use(cors.Handler(cors.Options{
			AllowedOrigins:     envVars.CORS.AllowedOrigins,
			AllowedMethods:     envVars.CORS.AllowedMethods,
			AllowedHeaders:     envVars.CORS.AllowedHeaders,
			ExposedHeaders:     envVars.CORS.ExposedHeaders,
			AllowCredentials:   envVars.CORS.AllowCredentials,
			MaxAge:             envVars.CORS.MaxAge,
			OptionsPassthrough: envVars.CORS.OptionsPassthrough,
			Debug:              logger.Enabled(context.TODO(), slog.LevelDebug),
		}))
	}

	return router
}

// ListenAndServe listens and serves the HTTP server.
func ListenAndServe(ctx context.Context, router *chi.Mux, envVars ServerConfig) error {
	router.Get(pathHealthz, func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			slog.Error("failed to write response: " + err.Error())
		}
	})

	serverErr := make(chan error, 1)

	// setup prometheus handler if enabled
	promServer, err := CreatePrometheusServer(router, envVars.Port)
	if err != nil {
		return err
	}

	if promServer != nil {
		defer func() {
			err := promServer.Shutdown(context.Background())
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Warn("failed to shutdown prometheus server: " + err.Error())
			}
		}()

		go func() {
			err := promServer.ListenAndServe()
			if !errors.Is(err, http.ErrServerClosed) {
				serverErr <- err
			}
		}()
	}

	maxHeaderBytes := http.DefaultMaxHeaderBytes

	if envVars.MaxHeaderKilobytes > 0 {
		maxHeaderBytes = envVars.MaxHeaderKilobytes * kilobyte
	}

	server := http.Server{
		Addr: fmt.Sprintf(":%d", envVars.Port),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		Handler:           router,
		ReadTimeout:       time.Duration(envVars.ReadTimeout),
		ReadHeaderTimeout: time.Duration(envVars.ReadHeaderTimeout),
		WriteTimeout:      time.Duration(envVars.WriteTimeout),
		IdleTimeout:       time.Duration(envVars.IdleTimeout),
		MaxHeaderBytes:    maxHeaderBytes,
	}

	go func() {
		var err error

		if envVars.TLSCertFile != "" || envVars.TLSKeyFile != "" {
			slog.Info("Listening server and serving TLS on " + server.Addr)
			err = server.ListenAndServeTLS(envVars.TLSCertFile, envVars.TLSKeyFile)
		} else {
			slog.Info("Listening server on " + server.Addr)
			err = server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for interruption.
	select {
	case err := <-serverErr:
		// Error when starting HTTP server.
		return err
	case <-ctx.Done():
		// Wait for first CTRL+C.
		slog.Info("received the quit signal, exiting...")
		// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
		return server.Shutdown(context.Background())
	}
}

// CreatePrometheusServer creates a Prometheus HTTP server from config.
func CreatePrometheusServer(router *chi.Mux, currentPort int) (*http.Server, error) {
	// setup prometheus handler if enabled
	prometheusEnabled := os.Getenv("OTEL_METRICS_EXPORTER") == "prometheus"
	prometheusPortEnv := os.Getenv("OTEL_EXPORTER_PROMETHEUS_PORT")

	if prometheusEnabled && prometheusPortEnv != "" {
		prometheusPort, err := strconv.Atoi(prometheusPortEnv)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errPrometheusInvalidPort, prometheusPortEnv)
		}

		if prometheusPort == currentPort {
			router.Handle(pathMetrics, promhttp.Handler())

			return nil, nil
		}

		promServer := createPrometheusServerInternal(prometheusPort)

		slog.Info(
			fmt.Sprintf("Listening prometheus server on %d", prometheusPort),
		)

		return promServer, nil
	}

	return nil, nil
}

func createPrometheusServerInternal(port int) *http.Server {
	mux := http.NewServeMux()
	mux.Handle(pathMetrics, promhttp.Handler())

	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}
}
