package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/hasura/gotel"
	"github.com/hasura/gotel/otelutils"
	"github.com/relychan/gohttps"
	"github.com/relychan/goutils"
)

func main() {
	os.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "1000")

	serverConfig, err := goutils.ReadJSONOrYAMLFile[gohttps.ServerConfig]("config.yaml")
	if err != nil {
		panic(err)
	}

	logger, _, err := otelutils.NewJSONLogger(serverConfig.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %s", err)
	}

	otlpConfig := &gotel.OTLPConfig{
		ServiceName:     "example",
		OtlpEndpoint:    "http://localhost:4317",
		OtlpProtocol:    gotel.OTLPProtocolGRPC,
		MetricsExporter: "otlp",
		LogsExporter:    "otlp",
	}

	ts, err := gotel.SetupOTelExporters(context.Background(), otlpConfig, "v0.1.0", logger)
	if err != nil {
		log.Fatal(err)
	}

	defer ts.Shutdown(context.TODO())

	router := gohttps.NewRouter(serverConfig, ts.Logger)
	router.Use(gotel.NewTracingMiddleware(ts, gotel.ResponseWriterWrapperFunc(func(w http.ResponseWriter, protoMajor int) gotel.WrapResponseWriter {
		return middleware.NewWrapResponseWriter(w, protoMajor)
	})))
	router.Use(middleware.AllowContentType("application/json"))

	router.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Hello world"))
	})

	err = gohttps.ListenAndServe(context.Background(), router, serverConfig)
	if err != nil {
		log.Fatalf("failed to serve http: %s", err)
	}
}
