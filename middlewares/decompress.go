// Copyright 2026 RelyChan Pte. Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middlewares

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/relychan/gocompress"
	"github.com/relychan/gohttps/httputils"
	"github.com/relychan/goutils"
	"github.com/relychan/goutils/httpheader"
)

// Decompress tries to decompress the request body if the Content-Encoding header is set.
// Responds with a 415 Unsupported Media Type status if the content type is not supported.
func Decompress(next http.Handler) http.Handler {
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
			decompressContentEncoding(w, r, next, requestEncodings[0])

			return
		}

		decompressContentEncodingSlice(w, r, next, requestEncodings)
	}

	return http.HandlerFunc(fn)
}

func decompressContentEncoding(
	w http.ResponseWriter,
	r *http.Request,
	next http.Handler,
	requestEncoding string,
) {
	logger := httputils.GetRequestLogger(r)

	formats, err := gocompress.DefaultCompressor.ParseSupportedEncoding(requestEncoding)
	if err != nil {
		logger.Warn(
			"error happened when parsing Content-Encoding",
			slog.String("error", err.Error()),
		)
	}

	if len(formats) == 0 {
		respondUnsupportedContentEncoding(w, r, logger)

		return
	}

	decompressedBody, err := gocompress.DefaultCompressor.DecompressFormat(
		r.Body,
		formats...,
	)
	if err != nil {
		respondDecompressionError(w, r, err, logger)

		return
	}

	r.Body = decompressedBody

	next.ServeHTTP(w, r)
}

func decompressContentEncodingSlice(
	w http.ResponseWriter,
	r *http.Request,
	next http.Handler,
	requestEncodings []string,
) {
	logger := httputils.GetRequestLogger(r)

	bodyBytes, err := io.ReadAll(r.Body)

	goutils.CatchWarnErrorFunc(r.Body.Close)

	if err != nil {
		respondDecompressionError(w, r, err, logger)

		return
	}

	bodyReader := bytes.NewReader(bodyBytes)
	isEncodingSupported := false

	var decompressErr error

	// All encodings in the request must be allowed
	for i, encoding := range requestEncodings {
		formats, err := gocompress.DefaultCompressor.ParseSupportedEncoding(encoding)
		if err != nil {
			logger.Warn(
				"error happened when parsing Content-Encoding",
				slog.String("error", err.Error()),
			)
		}

		if len(formats) == 0 {
			continue
		}

		isEncodingSupported = true

		if i > 0 {
			_, err := bodyReader.Seek(0, io.SeekStart)
			if err != nil {
				respondDecompressionError(w, r, err, logger)

				return
			}
		}

		decompressedBody, err := gocompress.DefaultCompressor.DecompressFormat(
			io.NopCloser(bodyReader),
			formats...,
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
		respondDecompressionError(w, r, decompressErr, logger)
	} else {
		respondUnsupportedContentEncoding(w, r, logger)
	}
}

func respondUnsupportedContentEncoding(
	w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
) {
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
		logger.Error(
			"failed to write response",
			slog.String("error", writeErr.Error()),
		)
	}
}

func respondDecompressionError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	logger *slog.Logger,
) {
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
		logger.Error(
			"failed to write response",
			slog.String("error", writeErr.Error()),
		)
	}
}
