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
//
// # This middleware replaces default encoders with klaupost
//
// Passing a compression level of 5 is sensible value.
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
