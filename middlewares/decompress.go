package middlewares

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/relychan/gocompress"
	"github.com/relychan/gohttps/httputils"
	"github.com/relychan/goutils"
)

// Decompress tries to decompress the request body if the Content-Encoding header is set.
// Responds with a 415 Unsupported Media Type status if the content type is not supported.
func Decompress(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil || r.Body == http.NoBody {
			next.ServeHTTP(w, r)

			return
		}

		requestEncodings := r.Header["Content-Encoding"]
		// skip check for empty content body or no Content-Encoding
		if r.ContentLength == 0 {
			next.ServeHTTP(w, r)

			return
		}

		// All encodings in the request must be allowed
		for _, encoding := range requestEncodings {
			trimmedEncoding := strings.TrimSpace(strings.ToLower(encoding))

			if gocompress.DefaultCompressor.IsEncodingSupported(trimmedEncoding) {
				decompressedBody, err := gocompress.DefaultCompressor.Decompress(
					r.Body,
					trimmedEncoding,
				)
				if err != nil {
					statusCode := http.StatusInternalServerError
					body := goutils.NewServerError(goutils.ErrorDetail{
						Detail: err.Error(),
					})
					body.Instance = r.URL.Path

					writeErr := httputils.WriteResponseJSON(w, statusCode, body)
					if writeErr != nil {
						httputils.GetRequestLogger(r).Error(
							"failed to write response",
							slog.String("error", writeErr.Error()),
						)
					}

					return
				}

				r.Body = decompressedBody

				next.ServeHTTP(w, r)

				return
			}
		}

		w.WriteHeader(http.StatusUnsupportedMediaType)
	}

	return http.HandlerFunc(fn)
}
