package middlewares

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/relychan/gohttps/httputils"
	"github.com/relychan/goutils"
)

// MaxBodySize creates a middleware with the max body size validation.
func MaxBodySize(maxBodySizeKilobytes int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if maxBodySizeKilobytes <= 0 {
			return next
		}

		maxBodySize := int64(maxBodySizeKilobytes * 1024)

		errorMessage := fmt.Sprintf("Request body size exceeded %d KB(s)", maxBodySizeKilobytes)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil || r.Body == http.NoBody {
				next.ServeHTTP(w, r)

				return
			}

			if r.ContentLength >= maxBodySize {
				statusCode := http.StatusRequestEntityTooLarge
				body := goutils.RFC9457Error{
					Type:     "about:blank",
					Title:    http.StatusText(statusCode),
					Detail:   errorMessage,
					Status:   statusCode,
					Code:     "413-01",
					Instance: r.URL.Path,
				}

				err := httputils.WriteResponseJSON(w, statusCode, body)
				if err != nil {
					httputils.GetRequestLogger(r).Error(
						"failed to write response",
						slog.String("error", err.Error()),
					)
				}

				return
			}

			// Always wrap the request body with MaxBytesReader to enforce the limit
			body := r.Body
			r.Body = http.MaxBytesReader(w, body, maxBodySize)

			next.ServeHTTP(w, r)
		})
	}
}
