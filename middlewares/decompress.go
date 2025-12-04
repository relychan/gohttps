package middlewares

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/relychan/gocompress"
	"github.com/relychan/gohttps/httputils"
	"github.com/relychan/goutils"
	"github.com/relychan/goutils/httpheader"
)

// Decompress tries to decompress the request body if the Content-Encoding header is set.
// Responds with a 415 Unsupported Media Type status if the content type is not supported.
func Decompress(next http.Handler) http.Handler { //nolint:funlen
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil || r.Body == http.NoBody {
			next.ServeHTTP(w, r)

			return
		}

		requestEncodings := r.Header[httpheader.ContentEncoding]
		// skip check for empty content body or no Content-Encoding
		if r.ContentLength == 0 || len(requestEncodings) == 0 {
			next.ServeHTTP(w, r)

			return
		}

		if len(requestEncodings) == 1 {
			trimmedEncoding := strings.TrimSpace(strings.ToLower(requestEncodings[0]))

			if !gocompress.DefaultCompressor.IsEncodingSupported(trimmedEncoding) {
				respondUnsupportedContentEncoding(w, r)

				return
			}

			decompressedBody, err := gocompress.DefaultCompressor.Decompress(
				r.Body,
				trimmedEncoding,
			)
			if err != nil {
				respondDecompressionError(w, r, err)

				return
			}

			r.Body = decompressedBody

			next.ServeHTTP(w, r)

			return
		}

		bodyBytes, err := io.ReadAll(r.Body)

		goutils.CatchWarnErrorFunc(r.Body.Close)

		if err != nil {
			respondDecompressionError(w, r, err)

			return
		}

		bodyReader := bytes.NewReader(bodyBytes)
		isEncodingSupported := false

		var decompressErr error

		// All encodings in the request must be allowed
		for i, encoding := range requestEncodings {
			trimmedEncoding := strings.TrimSpace(strings.ToLower(encoding))

			if !gocompress.DefaultCompressor.IsEncodingSupported(trimmedEncoding) {
				continue
			}

			isEncodingSupported = true

			if i > 0 {
				_, err := bodyReader.Seek(0, io.SeekStart)
				if err != nil {
					respondDecompressionError(w, r, err)

					return
				}
			}

			decompressedBody, err := gocompress.DefaultCompressor.Decompress(
				io.NopCloser(bodyReader),
				trimmedEncoding,
			)
			if err != nil {
				decompressErr = err

				continue
			}

			r.Body = decompressedBody

			next.ServeHTTP(w, r)

			return
		}

		if isEncodingSupported {
			respondDecompressionError(w, r, decompressErr)
		} else {
			respondUnsupportedContentEncoding(w, r)
		}
	}

	return http.HandlerFunc(fn)
}

func respondUnsupportedContentEncoding(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusUnsupportedMediaType
	body := goutils.NewRFC9457Error(
		statusCode,
		fmt.Sprintf(
			"Content-Encoding %v is unsupported",
			r.Header[httpheader.ContentEncoding],
		),
	)

	body.Instance = r.URL.Path

	writeErr := httputils.WriteResponseJSON(w, statusCode, body)
	if writeErr != nil {
		httputils.GetRequestLogger(r).Error(
			"failed to write response",
			slog.String("error", writeErr.Error()),
		)
	}
}

func respondDecompressionError(w http.ResponseWriter, r *http.Request, err error) {
	message := "failed to decompress body"

	if err != nil {
		message = err.Error()
	}

	body := goutils.NewBadRequestError(goutils.ErrorDetail{
		Detail: message,
	})
	body.Instance = r.URL.Path

	writeErr := httputils.WriteResponseJSON(w, body.Status, body)
	if writeErr != nil {
		httputils.GetRequestLogger(r).Error(
			"failed to write response",
			slog.String("error", writeErr.Error()),
		)
	}
}
