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

// Package middlewares defines custom middlewares for the chi router.
package middlewares

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
)

// Compress is a middleware that compresses response
// body of a given content types to a data format based
// on Accept-Encoding request header. It uses a given
// compression level.
// This middleware replaces default encoders with klauspost.
func Compress(level int, types ...string) func(next http.Handler) http.Handler {
	c := middleware.NewCompressor(level, types...)

	c.SetEncoder("deflate", encoderDeflate)
	c.SetEncoder("gzip", encoderGzip)
	c.SetEncoder("zstd", encoderZstd)

	return c.Handler
}

func encoderGzip(w io.Writer, level int) io.Writer {
	gw, err := gzip.NewWriterLevel(w, level)
	if err != nil {
		return nil
	}

	return gw
}

func encoderDeflate(w io.Writer, level int) io.Writer {
	dw, err := flate.NewWriter(w, level)
	if err != nil {
		return nil
	}

	return dw
}

func encoderZstd(w io.Writer, level int) io.Writer {
	dw, err := zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(level)))
	if err != nil {
		return nil
	}

	return dw
}
