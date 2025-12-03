package gohttps

import (
	"errors"

	"github.com/relychan/goutils"
)

const (
	pathMetrics = "/metrics"
	pathHealthz = "/healthz"
)

const (
	kilobyte = 1024
)

var errPrometheusInvalidPort = errors.New("invalid prometheus port")

// ServerConfig holds information of required environment variables.
type ServerConfig struct {
	Port               int              `env:"PORT" envDefault:"8080" json:"port,omitempty" yaml:"port,omitempty"`
	LogLevel           string           `env:"LOG_LEVEL" envDefault:"INFO" json:"logLevel,omitempty" yaml:"logLevel,omitempty"`
	CompressionLevel   int              `env:"SERVER_COMPRESSION_LEVEL" json:"compressionLevel,omitempty" yaml:"compressionLevel,omitempty"`
	RequestTimeout     goutils.Duration `env:"SERVER_REQUEST_TIMEOUT" json:"requestTimeout,omitempty" yaml:"requestTimeout,omitempty"`
	ReadTimeout        goutils.Duration `env:"SERVER_READ_TIMEOUT" json:"readTimeout,omitempty" yaml:"readTimeout,omitempty"`
	ReadHeaderTimeout  goutils.Duration `env:"SERVER_READ_HEADER_TIMEOUT"  json:"readHeaderTimeout,omitempty" yaml:"readHeaderTimeout,omitempty"`
	WriteTimeout       goutils.Duration `env:"SERVER_WRITE_TIMEOUT" json:"writeTimeout,omitempty" yaml:"writeTimeout,omitempty"`
	IdleTimeout        goutils.Duration `env:"SERVER_IDLE_TIMEOUT" json:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty"`
	MaxHeaderKilobytes int              `env:"SERVER_MAX_HEADER_KILOBYTES" json:"maxHeaderKilobytes,omitempty" yaml:"maxHeaderKilobytes,omitempty"`
	MaxBodyKilobytes   int              `env:"SERVER_MAX_BODY_KILOBYTES" json:"maxBodyKilobytes,omitempty" yaml:"maxBodyKilobytes,omitempty"`
	TLSCertFile        string           `env:"SERVER_TLS_CERT_FILE" json:"tlsCertFile,omitempty" yaml:"tlsCertFile,omitempty"`
	TLSKeyFile         string           `env:"SERVER_TLS_KEY_FILE" json:"tlsKeyFile,omitempty" yaml:"tlsKeyFile,omitempty"`
	CORS               *CORSConfig      `json:"cors,omitempty" yaml:"cors,omitempty"`
}

// CORSConfig represents configurations of CORS.
type CORSConfig struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// An origin may contain a wildcard (*) to replace 0 or more characters
	// (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penalty.
	// Only one wildcard can be used per origin.
	// CORS is disabled if empty.
	AllowedOrigins []string `env:"SERVER_CORS_ALLOWED_ORIGINS" json:"allowedOrigins" yaml:"allowedOrigins" jsonschema:"nullable"`
	// AllowedMethods is a list of methods the client is allowed to use with cross-domain requests.
	// Default value is simple methods (HEAD, GET and POST).
	AllowedMethods []string `env:"SERVER_CORS_ALLOWED_METHODS" json:"allowedMethods" yaml:"allowedMethods" jsonschema:"nullable"`
	// AllowedHeaders is list of non simple headers the client is allowed to use with cross-domain requests.
	// If the special "*" value is present in the list, all headers will be allowed.
	// Default value is [] but "Origin" is always appended to the list.
	AllowedHeaders []string `env:"SERVER_CORS_ALLOWED_HEADERS" json:"allowedHeaders" yaml:"allowedHeaders" jsonschema:"nullable"`
	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS API specification
	ExposedHeaders []string `env:"SERVER_CORS_EXPOSED_HEADERS" json:"exposedHeaders" yaml:"exposedHeaders" jsonschema:"nullable"`
	// AllowCredentials indicates whether the request can include user credentials like cookies,
	// HTTP authentication or client side SSL certificates.
	AllowCredentials bool `env:"SERVER_CORS_ALLOW_CREDENTIALS" json:"allowCredentials" yaml:"allowCredentials" jsonschema:"nullable"`
	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached
	MaxAge int `env:"SERVER_CORS_MAX_AGE" json:"maxAge" yaml:"maxAge" jsonschema:"nullable"`
	// OptionsPassthrough instructs preflight to let other potential next handlers to process the OPTIONS method.
	// Turn this on if your application handles OPTIONS.
	OptionsPassthrough bool `env:"SERVER_CORS_OPTIONS_PASSTHROUGH" json:"optionsPassthrough" yaml:"optionsPassthrough" jsonschema:"nullable"`
}
