package gohttps

import (
	"encoding/json"
	"errors"
	"fmt"
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
			`{"message": "%s"}`,
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

// DecodeRequestBody decodes the request body.
func DecodeRequestBody[T any](w http.ResponseWriter, r *http.Request, span trace.Span) (*T, bool) {
	if r.Body == nil {
		message := "request body is required"
		span.SetStatus(codes.Error, message)

		wErr := WriteResponseJSON(w, http.StatusUnprocessableEntity, map[string]string{
			"message": message,
		})
		if wErr != nil {
			SetWriteResponseErrorAttribute(span, wErr)
		}

		return nil, false
	}

	var input T

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		span.SetStatus(codes.Error, "failed to decode json")
		span.RecordError(err)

		wErr := WriteResponseJSON(w, http.StatusUnprocessableEntity, map[string]string{
			"message": err.Error(),
		})
		if wErr != nil {
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
		return value, RFC9457Error{
			Status:   http.StatusUnprocessableEntity,
			Title:    http.StatusText(http.StatusUnprocessableEntity),
			Detail:   "Failed to parse URL parameter",
			Instance: r.URL.Path,
			Errors: []ErrorDetail{
				{
					Detail:    "Invalid UUID format",
					Parameter: param,
				},
			},
		}
	}

	return value, nil
}

// GetURLParamInt64 gets a URL parameter and parses it as integer.
func GetURLParamInt64(r *http.Request, param string) (int64, error) {
	rawValue := chi.URLParam(r, param)

	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		return value, RFC9457Error{
			Status:   http.StatusUnprocessableEntity,
			Title:    http.StatusText(http.StatusUnprocessableEntity),
			Detail:   "Failed to parse URL parameter",
			Instance: r.URL.Path,
			Errors: []ErrorDetail{
				{
					Detail:    "Invalid integer format",
					Parameter: param,
				},
			},
		}
	}

	return value, nil
}
