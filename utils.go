package gohttps

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// WriteResponseJSON writes response data with json encode. Returns the response size.
func WriteResponseJSON(w http.ResponseWriter, statusCode int, body any) error {
	if body == nil {
		w.WriteHeader(statusCode)

		return nil
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSON)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	w.WriteHeader(statusCode)

	err := enc.Encode(body)
	if err != nil {
		_, wErr := fmt.Fprintf(
			w,
			`{"title": "%s"}`,
			http.StatusText(http.StatusInternalServerError),
		)
		if wErr != nil {
			return errors.Join(err, wErr)
		}

		return err
	}

	return nil
}

// WriteResponseError responds the error to the client.
func WriteResponseError(w http.ResponseWriter, err error) error {
	var httpError RFC9457Error

	statusCode := http.StatusInternalServerError

	if errors.As(err, &httpError) {
		if httpError.Status > 0 {
			statusCode = httpError.Status
		}

		return WriteResponseJSON(w, statusCode, httpError)
	}

	httpError.Status = statusCode
	httpError.Title = http.StatusText(statusCode)
	httpError.Detail = err.Error()

	return WriteResponseJSON(w, statusCode, httpError)
}

// DecodeRequestBody attempts to decode the HTTP request body into a value of type T.
//
// Type Parameters:
//
//	T - The type into which the request body should be decoded. Typically a struct matching the expected JSON payload.
//
// Parameters:
//
//	w    - The http.ResponseWriter used to write error responses if decoding fails.
//	r    - The *http.Request containing the body to decode.
//	span - The trace.Span used for recording tracing information and errors.
//
// Returns:
//
//	*T    - Pointer to the decoded value of type T, or nil if decoding fails.
//	bool  - True if decoding was successful, false otherwise.
//
// Behavior:
//   - If the request body is missing (nil or http.NoBody), the function writes an HTTP error response
//     (status 422 Unprocessable Entity) with message "request body is required".
//   - If the request body cannot be decoded as JSON into type T, the function writes an HTTP error response
//     (status 422 Unprocessable Entity) with message "Invalid request body".
//   - In both error cases, sets the span status to error and returns (nil, false).
//   - On success, returns a pointer to the decoded value and true.
//   - The function may write HTTP responses in case of error, but does not write on success.
func DecodeRequestBody[T any](
	w http.ResponseWriter,
	r *http.Request,
	span trace.Span,
) (*T, bool) {
	if r.Body == nil || r.Body == http.NoBody {
		message := "request body is required"
		span.SetStatus(codes.Error, message)

		respError := NewMissingBodyPropertyError(ErrorDetail{
			Detail:  "Request body is required",
			Pointer: "#",
		})

		wErr := WriteResponseJSON(w, http.StatusUnprocessableEntity, respError)
		if wErr != nil {
			logger := getRequestLogger(r)
			logger.Error("failed to write response", slog.String("error", wErr.Error()))
			SetWriteResponseErrorAttribute(span, wErr)
		}

		return nil, false
	}

	var input T

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		span.SetStatus(codes.Error, "failed to decode JSON")
		span.RecordError(err)

		logger := getRequestLogger(r)
		logger.Debug("failed to decode JSON", slog.String("error", err.Error()))

		respError := ErrBadRequest

		wErr := WriteResponseJSON(w, http.StatusUnprocessableEntity, respError)
		if wErr != nil {
			logger.Error("failed to write response", slog.String("error", wErr.Error()))
			SetWriteResponseErrorAttribute(span, wErr)
		}

		return nil, false
	}

	return &input, true
}

// SetWriteResponseErrorAttribute sets the error that happens when writing the HTTP response.
func SetWriteResponseErrorAttribute(span trace.Span, err error) {
	span.SetAttributes(attribute.String("http.response.write_error", err.Error()))
}

// GetURLParamUUID gets a URL parameter and parses it as UUID.
func GetURLParamUUID(r *http.Request, param string) (uuid.UUID, error) {
	rawValue := chi.URLParam(r, param)

	value, err := uuid.Parse(rawValue)
	if err != nil {
		respError := NewInvalidRequestHeaderFormatError(ErrorDetail{
			Detail:    "Invalid UUID format",
			Parameter: param,
		})
		respError.Instance = r.URL.Path

		return value, respError
	}

	return value, nil
}

// GetURLParamInt64 gets a URL parameter and parses it as integer.
func GetURLParamInt64(r *http.Request, param string) (int64, error) {
	rawValue := chi.URLParam(r, param)

	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		respError := NewInvalidRequestHeaderFormatError(ErrorDetail{
			Detail:    "Invalid integer format",
			Parameter: param,
		})
		respError.Instance = r.URL.Path

		return value, respError
	}

	return value, nil
}

func getRequestLogger(r *http.Request) *slog.Logger {
	return slog.Default().With(slog.String("request_id", getRequestID(r)))
}

func getRequestID(r *http.Request) string {
	requestID := r.Header.Get("x-request-id")
	if requestID != "" {
		return requestID
	}

	spanContext := trace.SpanContextFromContext(r.Context())
	if spanContext.HasTraceID() {
		return spanContext.TraceID().String()
	}

	return uuid.NewString()
}
