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

var (
	errPrometheusInvalidPort = errors.New("invalid prometheus port")
	errServerConfigRequired  = errors.New("server config is required")
)

// ServerConfig holds information of required environment variables.
type ServerConfig struct {
	// The port where the server is listening to.
	Port int `env:"PORT" envDefault:"8080" json:"port,omitempty" yaml:"port,omitempty"`
	// Level of the logger.
	LogLevel string `env:"LOG_LEVEL" envDefault:"INFO" json:"logLevel,omitempty" yaml:"logLevel,omitempty" jsonschema:"enum=INFO,enum=DEBUG,enum=WARN,enum=ERROR"`
	// Default level which the server uses to compress response bodies.
	CompressionLevel *int `env:"SERVER_COMPRESSION_LEVEL" json:"compressionLevel,omitempty" yaml:"compressionLevel,omitempty" jsonschema:"min=-1,max=9"`
	// The default timeout of every request. Return a 504 Gateway Timeout error to the client.
	RequestTimeout goutils.Duration `env:"SERVER_REQUEST_TIMEOUT" json:"requestTimeout,omitempty" yaml:"requestTimeout,omitempty"`
	// The maximum duration for reading the entire request, including the body.
	// A zero or negative value means there will be no timeout.
	ReadTimeout goutils.Duration `env:"SERVER_READ_TIMEOUT" json:"readTimeout,omitempty" yaml:"readTimeout,omitempty"`
	// The amount of time allowed to read request headers.
	// The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body.
	// If zero, the value of ReadTimeout is used. If negative, or if zero and ReadTimeout is zero or negative, there is no timeout.
	ReadHeaderTimeout goutils.Duration `env:"SERVER_READ_HEADER_TIMEOUT" json:"readHeaderTimeout,omitempty" yaml:"readHeaderTimeout,omitempty"`
	// The maximum duration before timing out writes of the response. It is reset whenever a new request's header is read.
	// Like ReadTimeout, it does not let Handlers make decisions on a per-request basis.
	// A zero or negative value means there will be no timeout.
	WriteTimeout goutils.Duration `env:"SERVER_WRITE_TIMEOUT" json:"writeTimeout,omitempty" yaml:"writeTimeout,omitempty"`
	// The maximum amount of time to wait for the next request when keep-alives are enabled.
	// If zero, the value of ReadTimeout is used.
	// If negative, or if zero and ReadTimeout is zero or negative, there is no timeout.
	IdleTimeout goutils.Duration `env:"SERVER_IDLE_TIMEOUT" json:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty"`
	// The maximum number of bytes the server will read parsing the request header's keys and values, including the request line.
	// It does not limit the size of the request body. If zero, DefaultMaxHeaderBytes is used.
	MaxHeaderKilobytes int `env:"SERVER_MAX_HEADER_KILOBYTES" json:"maxHeaderKilobytes,omitempty" yaml:"maxHeaderKilobytes,omitempty"`
	// The maximum number of bytes the server will read parsing the request body.
	// A zero or negative value means there will be no limit.
	MaxBodyKilobytes int `env:"SERVER_MAX_BODY_KILOBYTES" json:"maxBodyKilobytes,omitempty" yaml:"maxBodyKilobytes,omitempty"`
	// The TLS certificate file to enable TLS connections.
	TLSCertFile string `env:"SERVER_TLS_CERT_FILE" json:"tlsCertFile,omitempty" yaml:"tlsCertFile,omitempty"`
	// The TLS key file to enable TLS connections.
	TLSKeyFile string `env:"SERVER_TLS_KEY_FILE" json:"tlsKeyFile,omitempty" yaml:"tlsKeyFile,omitempty"`
	// The configuration container to setup the CORS middleware.
	CORS *CORSConfig `json:"cors,omitempty" yaml:"cors,omitempty"`
}

// CORSConfig represents configurations of CORS.
type CORSConfig struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// An origin may contain a wildcard (*) to replace 0 or more characters
	// (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penalty.
	// Only one wildcard can be used per origin.
	// CORS is disabled if empty.
	AllowedOrigins []string `env:"SERVER_CORS_ALLOWED_ORIGINS" json:"allowedOrigins,omitempty" yaml:"allowedOrigins,omitempty"`
	// AllowedMethods is a list of methods the client is allowed to use with cross-domain requests.
	// Default value is simple methods (HEAD, GET and POST).
	AllowedMethods []string `env:"SERVER_CORS_ALLOWED_METHODS" json:"allowedMethods,omitempty" yaml:"allowedMethods,omitempty"`
	// AllowedHeaders is list of non simple headers the client is allowed to use with cross-domain requests.
	// If the special "*" value is present in the list, all headers will be allowed.
	// Default value is [] but "Origin" is always appended to the list.
	AllowedHeaders []string `env:"SERVER_CORS_ALLOWED_HEADERS" json:"allowedHeaders,omitempty" yaml:"allowedHeaders,omitempty"`
	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS API specification
	ExposedHeaders []string `env:"SERVER_CORS_EXPOSED_HEADERS" json:"exposedHeaders,omitempty" yaml:"exposedHeaders,omitempty"`
	// AllowCredentials indicates whether the request can include user credentials like cookies,
	// HTTP authentication or client side SSL certificates.
	AllowCredentials bool `env:"SERVER_CORS_ALLOW_CREDENTIALS" json:"allowCredentials,omitempty" yaml:"allowCredentials"`
	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached
	MaxAge int `env:"SERVER_CORS_MAX_AGE" json:"maxAge,omitempty" yaml:"maxAge,omitempty" jsonschema:"min=0"`
	// OptionsPassthrough instructs preflight to let other potential next handlers to process the OPTIONS method.
	// Turn this on if your application handles OPTIONS.
	OptionsPassthrough bool `env:"SERVER_CORS_OPTIONS_PASSTHROUGH" json:"optionsPassthrough,omitempty" yaml:"optionsPassthrough,omitempty"`
}
