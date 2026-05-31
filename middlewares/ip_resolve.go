package middlewares

import (
	"errors"
	"net/http"
	"net/netip"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
)

var (
	errClientIPHeaderRequired = errors.New(
		"headers are required for the header client ip resolution type",
	)
	errClientIPHeaderEmpty               = errors.New("header must be a non-empty string")
	errClientIPTrustedIPPrefixesRequired = errors.New("trusted IP prefixes must not be empty")
	errClientIPInvalidNumTrustedProxies  = errors.New(
		"the number of trusted proxies must be larger than 0",
	)
)

// ClientIPResolutionType represents the enum of IP resolution strategy types.
type ClientIPResolutionType string

const (
	// ClientIPFromRemoteAddr stores the client IP read from the TCP RemoteAddr of the incoming request — the IP address of whoever opened the connection to this server.
	// Use this strategy when this server is directly connected to the public internet with NO reverse proxy in front of it.
	// Behind a reverse proxy, RemoteAddr is the proxy's IP, not the client's — use ClientIPFromHeader or ClientIPFromXFF instead.
	// IPv4 clients on a dual-stack listener surface as ::ffff:a.b.c.d; they fold to plain v4 before storage so one logical client maps to one key.
	// IPv6 zones are preserved (link-local connections may legitimately have one).
	ClientIPFromRemoteAddr ClientIPResolutionType = "remote_addr"
	// ClientIPFromHeader stores the client IP from a single-IP header set by your reverse proxy. Read it with GetClientIP.
	// Only safe with headers your proxy unconditionally OVERWRITES on every request, e.g.:  X-Real-IP, X-Client-IP, CF-Connecting-IP — Cloudflare.
	// If the header reaches us with multiple values (misconfigured proxy that appends, or a downstream proxy not stripping a client-supplied value),
	// the LAST value wins — that's the one set by the hop closest to us, and therefore the most trusted.
	// Fail-closed if the last value doesn't parse: no client IP is set rather than falling back to earlier (less-trusted) values.
	// v4-mapped IPv6 (::ffff:a.b.c.d) folds to plain v4 and IPv6 zones are stripped before storage.
	ClientIPFromHeader ClientIPResolutionType = "header"
	// ClientIPFromXForwardedFor stores the client IP read from the X-Forwarded-For header, walking the chain right-to-left and skipping any IP that falls within one of the given trusted CIDR prefixes.
	// The first IP that is not trusted is the client.
	ClientIPFromXForwardedFor ClientIPResolutionType = "x_forward_for"
	// ClientIPFromXForwardForTrustedProxies stores the client IP read from the X-Forwarded-For header, given the exact number of trusted reverse proxies between this server and the public internet.
	// It returns the IP at position len(xff) - numTrustedProxies in the merged X-Forwarded-For list — the IP added by the outermost of your trusted proxies,
	// the only IP in the chain that none of your proxies have allowed an attacker to forge.
	ClientIPFromXForwardForTrustedProxies ClientIPResolutionType = "x_forward_for_trusted_proxies"
)

// ClientIPConfig represents the configuration for IP resolution from HTTP requests.
type ClientIPConfig struct {
	// Type of the strategy that the client IP should be parsed from.
	Type ClientIPResolutionType `env:"SERVER_CLIENT_IP_RESOLUTION_TYPE" json:"type" yaml:"type"`
	// List of headers to be looked up. Required if type=remote_addr.
	Headers []string `env:"SERVER_CLIENT_IP_HEADERS" json:"headers,omitempty" yaml:"headers,omitempty"`
	// List of CIDR prefixes to be trusted when parsing the client IP from the X-Forwarded-For header. Required if type=x_forward_for.
	TrustedIPPrefixes []string `env:"SERVER_TRUSTED_CLIENT_IP_PREFIXES" json:"trustedIpPrefixes,omitempty" yaml:"trustedIpPrefixes,omitempty"`
	// The exact number of trusted reverse proxies between this server and the public internet. Required if type=x_forward_for_trusted_proxies.
	NumTrustedProxies int `env:"SERVER_CLIENT_IP_NUM_TRUSTED_PROXIES" json:"numTrustedProxies,omitempty" yaml:"numTrustedProxies,omitempty"`
}

// Validate checks if the configuration is valid.
func (cic ClientIPConfig) Validate() error {
	switch cic.Type {
	case ClientIPFromHeader:
		if len(cic.Headers) == 0 {
			return errClientIPHeaderRequired
		}

		for _, header := range cic.Headers {
			if strings.TrimSpace(header) == "" {
				return errClientIPHeaderEmpty
			}
		}

		return nil
	case ClientIPFromXForwardedFor:
		if len(cic.TrustedIPPrefixes) == 0 {
			return errClientIPTrustedIPPrefixesRequired
		}

		return nil
	case ClientIPFromXForwardForTrustedProxies:
		if cic.NumTrustedProxies == 0 {
			return errClientIPInvalidNumTrustedProxies
		}

		for _, p := range cic.TrustedIPPrefixes {
			_, err := netip.ParsePrefix(p)
			if err != nil {
				return err
			}
		}

		return nil
	case ClientIPFromRemoteAddr:
		fallthrough
	default:
		return nil
	}
}

// ClientIP stores the client IP from a single-IP header set by your reverse proxy.
// Read it with middleware.GetClientIP function of the chi router.
func ClientIP(config *ClientIPConfig) func(next http.Handler) http.Handler {
	if config == nil {
		return middleware.ClientIPFromRemoteAddr
	}

	switch config.Type {
	case ClientIPFromHeader:
		if len(config.Headers) == 1 {
			return middleware.ClientIPFromHeader(config.Headers[0])
		}

		return clientIPFromHeaders(config.Headers)
	case ClientIPFromXForwardedFor:
		return middleware.ClientIPFromXFF(config.TrustedIPPrefixes...)
	case ClientIPFromXForwardForTrustedProxies:
		return middleware.ClientIPFromXFFTrustedProxies(config.NumTrustedProxies)
	case ClientIPFromRemoteAddr:
		fallthrough
	default:
		return middleware.ClientIPFromRemoteAddr
	}
}

func clientIPFromHeaders(headers []string) func(next http.Handler) http.Handler {
	for i := range headers {
		headers[i] = http.CanonicalHeaderKey(headers[i])
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, header := range headers {
				_, ok := r.Header[header]
				if ok {
					middleware.ClientIPFromHeader(header)

					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
