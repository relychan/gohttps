package gohttps

import (
	"errors"
	"time"
)

const (
	pathMetrics = "/metrics"
	pathHealthz = "/healthz"
)

const (
	kilobyte = 1024
)

var errPrometheusInvalidPort = errors.New("invalid prometheus port")

const (
	// ContentEncodingHeader is the constant for the Content-Encoding header name.
	ContentEncodingHeader = "Content-Encoding"
	// ContentTypeHeader is the constant for the Content-Type header name.
	ContentTypeHeader = "Content-Type"
	// ContentTypeJSON is the constant for the application/json content type.
	ContentTypeJSON = "application/json"
	// ContentTypeNdJSON is the constant for the application/x-ndjson content type.
	ContentTypeNdJSON = "application/x-ndjson"
	// ContentTypeXML is the constant for the application/xml content type.
	ContentTypeXML = "application/xml"
	// ContentTypeFormURLEncoded is the constant for the application/x-www-form-urlencoded content type.
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
	// ContentTypeMultipartFormData is the constant for the multipart/form-data content type.
	ContentTypeMultipartFormData = "multipart/form-data"
	// ContentTypeTextPlain is the constant for the text/plain content type.
	ContentTypeTextPlain = "text/plain"
	// ContentTypeTextHTML is the constant for the text/html content type.
	ContentTypeTextHTML = "text/html"
	// ContentTypeTextXML is the constant for the text/xml content type.
	ContentTypeTextXML = "text/xml"
	// ContentTypeOctetStream is the constant for the application/octet-stream content type.
	ContentTypeOctetStream = "application/octet-stream"
)

// ServerConfig holds information of required environment variables.
type ServerConfig struct {
	Port               int           `env:"PORT"                        envDefault:"8080"`
	LogLevel           string        `env:"LOG_LEVEL"                   envDefault:"INFO"`
	CompressionLevel   int           `env:"SERVER_COMPRESSION_LEVEL"`
	RequestTimeout     time.Duration `env:"SERVER_REQUEST_TIMEOUT"`
	ReadTimeout        time.Duration `env:"SERVER_READ_TIMEOUT"`
	ReadHeaderTimeout  time.Duration `env:"SERVER_READ_HEADER_TIMEOUT"`
	WriteTimeout       time.Duration `env:"SERVER_WRITE_TIMEOUT"`
	IdleTimeout        time.Duration `env:"SERVER_IDLE_TIMEOUT"`
	MaxHeaderKilobytes int           `env:"SERVER_MAX_HEADER_KILOBYTES"`
	MaxBodyKilobytes   int           `env:"SERVER_MAX_BODY_KILOBYTES"`
	TLSCertFile        string        `env:"SERVER_TLS_CERT_FILE"`
	TLSKeyFile         string        `env:"SERVER_TLS_KEY_FILE"`

	// CorsAllowedOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// An origin may contain a wildcard (*) to replace 0 or more characters
	// (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penalty.
	// Only one wildcard can be used per origin.
	// CORS is disabled if empty.
	CorsAllowedOrigins []string `env:"SERVER_CORS_ALLOWED_ORIGINS"`
	// CorsAllowedMethods is a list of methods the client is allowed to use with cross-domain requests.
	// Default value is simple methods (HEAD, GET and POST).
	CorsAllowedMethods []string `env:"SERVER_CORS_ALLOWED_METHODS"`
	// CorsAllowedHeaders is list of non simple headers the client is allowed to use with cross-domain requests.
	// If the special "*" value is present in the list, all headers will be allowed.
	// Default value is [] but "Origin" is always appended to the list.
	CorsAllowedHeaders []string `env:"SERVER_CORS_ALLOWED_HEADERS"`
	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS API specification
	CorsExposedHeaders []string `env:"SERVER_CORS_EXPOSED_HEADERS"`
	// CorsAllowCredentials indicates whether the request can include user credentials like cookies,
	// HTTP authentication or client side SSL certificates.
	CorsAllowCredentials bool `env:"SERVER_CORS_ALLOW_CREDENTIALS"`
	// CorsMaxAge indicates how long (in seconds) the results of a preflight request can be cached
	CorsMaxAge int `env:"SERVER_CORS_MAX_AGE"`
	// CorsOptionsPassthrough instructs preflight to let other potential next handlers to process the OPTIONS method.
	// Turn this on if your application handles OPTIONS.
	CorsOptionsPassthrough bool `env:"SERVER_CORS_OPTIONS_PASSTHROUGH"`
}

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/http.
type contextKey struct {
	name string
}

// NewContextKey creates a new context key.
func NewContextKey(name string) *contextKey {
	return &contextKey{name}
}

func (k *contextKey) String() string {
	return "context value " + k.name
}
