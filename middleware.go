package gohttps

import (
	"fmt"
	"net/http"

	"github.com/relychan/goutils"
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
			if r.Body != nil {
				if r.ContentLength >= maxBodySize {
					err := goutils.RFC9457Error{
						Type:     "about:blank",
						Title:    http.StatusText(http.StatusRequestEntityTooLarge),
						Detail:   errorMessage,
						Status:   http.StatusRequestEntityTooLarge,
						Code:     "413-01",
						Instance: r.URL.Path,
					}

					wErr := WriteResponseJSON(w, http.StatusRequestEntityTooLarge, err)
					if wErr != nil {
						span := trace.SpanFromContext(r.Context())
						SetWriteResponseErrorAttribute(span, wErr)
					}

					return
				}

				// Always wrap the request body with MaxBytesReader to enforce the limit
				body := r.Body
				r.Body = http.MaxBytesReader(w, body, maxBodySize)
			}

			next.ServeHTTP(w, r)
		})
	}
}
