package gohttps

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// MaxBodySizeMiddleware creates a middleware with logger context.
func MaxBodySizeMiddleware(maxBodySizeKilobytes int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if maxBodySizeKilobytes <= 0 {
			return next
		}

		maxBodySize := int64(maxBodySizeKilobytes * kilobyte)

		errorMessage := fmt.Sprintf("Request body size exceeded %d KB(s)", maxBodySizeKilobytes)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength >= maxBodySize {
				err := NewRFC9457Error(http.StatusRequestEntityTooLarge, errorMessage)
				err.Instance = r.URL.Path

				wErr := WriteResponseJSON(w, http.StatusRequestEntityTooLarge, err)
				if wErr != nil {
					span := trace.SpanFromContext(r.Context())
					SetWriteResponseErrorAttribute(span, wErr)
				}

				return
			}

			// if the content length is unknown, wrap the request body with MaxBytesReader
			if r.ContentLength <= -1 {
				body := r.Body
				r.Body = http.MaxBytesReader(w, body, maxBodySize)
			}

			next.ServeHTTP(w, r)
		})
	}
}
