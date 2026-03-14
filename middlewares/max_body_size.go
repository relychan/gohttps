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

		errorMessage := fmt.Sprintf("Request body size exceeded %d KB", maxBodySizeKilobytes)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil || r.Body == http.NoBody {
				next.ServeHTTP(w, r)

				return
			}

			if r.ContentLength > maxBodySize {
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
