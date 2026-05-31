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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
)

// captureClientIP is a handler that records the client IP resolved by the middleware.
func captureClientIP(t *testing.T, got *string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got = middleware.GetClientIP(r.Context())
		w.WriteHeader(http.StatusOK)
	})
}

// ---- ClientIPConfig.Validate ------------------------------------------------

func TestClientIPConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ClientIPConfig
		wantErr error
	}{
		{
			name:    "remote_addr is always valid",
			config:  ClientIPConfig{Type: ClientIPFromRemoteAddr},
			wantErr: nil,
		},
		{
			name:    "empty type defaults to valid",
			config:  ClientIPConfig{},
			wantErr: nil,
		},
		{
			name:    "header valid with one header",
			config:  ClientIPConfig{Type: ClientIPFromHeader, Headers: []string{"X-Real-IP"}},
			wantErr: nil,
		},
		{
			name:    "header valid with multiple headers",
			config:  ClientIPConfig{Type: ClientIPFromHeader, Headers: []string{"X-Real-IP", "CF-Connecting-IP"}},
			wantErr: nil,
		},
		{
			name:    "header missing headers list",
			config:  ClientIPConfig{Type: ClientIPFromHeader},
			wantErr: errClientIPHeaderRequired,
		},
		{
			name:    "header with blank header entry",
			config:  ClientIPConfig{Type: ClientIPFromHeader, Headers: []string{"  "}},
			wantErr: errClientIPHeaderEmpty,
		},
		{
			name:    "header with empty string entry",
			config:  ClientIPConfig{Type: ClientIPFromHeader, Headers: []string{"X-Real-IP", ""}},
			wantErr: errClientIPHeaderEmpty,
		},
		{
			name:    "x_forward_for valid",
			config:  ClientIPConfig{Type: ClientIPFromXForwardedFor, TrustedIPPrefixes: []string{"10.0.0.0/8"}},
			wantErr: nil,
		},
		{
			name:    "x_forward_for missing trusted prefixes",
			config:  ClientIPConfig{Type: ClientIPFromXForwardedFor},
			wantErr: errClientIPTrustedIPPrefixesRequired,
		},
		{
			name:    "x_forward_for_trusted_proxies valid",
			config:  ClientIPConfig{Type: ClientIPFromXForwardForTrustedProxies, NumTrustedProxies: 1},
			wantErr: nil,
		},
		{
			name:    "x_forward_for_trusted_proxies zero proxies",
			config:  ClientIPConfig{Type: ClientIPFromXForwardForTrustedProxies, NumTrustedProxies: 0},
			wantErr: errClientIPInvalidNumTrustedProxies,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()

			// For the invalid-CIDR case we only check that an error is returned.
			if tc.name == "x_forward_for_trusted_proxies invalid CIDR" {
				if err == nil {
					t.Fatal("expected error for invalid CIDR prefix, got nil")
				}
				return
			}

			if err != tc.wantErr {
				t.Errorf("Validate() = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// ---- ClientIP middleware -----------------------------------------------------

func TestClientIP_NilConfig(t *testing.T) {
	var got string
	mw := ClientIP(nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.5:1234"
	w := httptest.NewRecorder()

	mw(captureClientIP(t, &got)).ServeHTTP(w, req)

	if got != "203.0.113.5" {
		t.Errorf("expected 203.0.113.5, got %q", got)
	}
}

func TestClientIP_RemoteAddr(t *testing.T) {
	var got string
	mw := ClientIP(&ClientIPConfig{Type: ClientIPFromRemoteAddr})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.7:4567"
	w := httptest.NewRecorder()

	mw(captureClientIP(t, &got)).ServeHTTP(w, req)

	if got != "198.51.100.7" {
		t.Errorf("expected 198.51.100.7, got %q", got)
	}
}

func TestClientIP_RemoteAddr_IPv4MappedIPv6(t *testing.T) {
	var got string
	mw := ClientIP(&ClientIPConfig{Type: ClientIPFromRemoteAddr})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "[::ffff:192.0.2.1]:9000"
	w := httptest.NewRecorder()

	mw(captureClientIP(t, &got)).ServeHTTP(w, req)

	// v4-mapped must fold to plain v4
	if got != "192.0.2.1" {
		t.Errorf("expected 192.0.2.1, got %q", got)
	}
}

func TestClientIP_SingleHeader(t *testing.T) {
	var got string
	mw := ClientIP(&ClientIPConfig{
		Type:    ClientIPFromHeader,
		Headers: []string{"X-Real-IP"},
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "203.0.113.42")
	w := httptest.NewRecorder()

	mw(captureClientIP(t, &got)).ServeHTTP(w, req)

	if got != "203.0.113.42" {
		t.Errorf("expected 203.0.113.42, got %q", got)
	}
}

func TestClientIP_SingleHeader_LastValueWins(t *testing.T) {
	var got string
	mw := ClientIP(&ClientIPConfig{
		Type:    ClientIPFromHeader,
		Headers: []string{"X-Real-IP"},
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Two values — last is most trusted (closest hop).
	req.Header.Add("X-Real-IP", "1.2.3.4")
	req.Header.Add("X-Real-IP", "203.0.113.99")
	w := httptest.NewRecorder()

	mw(captureClientIP(t, &got)).ServeHTTP(w, req)

	if got != "203.0.113.99" {
		t.Errorf("expected 203.0.113.99 (last value), got %q", got)
	}
}

func TestClientIP_SingleHeader_CaseInsensitive(t *testing.T) {
	var got string
	mw := ClientIP(&ClientIPConfig{
		Type:    ClientIPFromHeader,
		Headers: []string{"x-real-ip"}, // lowercase — should be canonicalized
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "10.1.2.3")
	w := httptest.NewRecorder()

	mw(captureClientIP(t, &got)).ServeHTTP(w, req)

	if got != "10.1.2.3" {
		t.Errorf("expected 10.1.2.3, got %q", got)
	}
}

func TestClientIP_MultipleHeaders_FirstMatch(t *testing.T) {
	var got string
	mw := ClientIP(&ClientIPConfig{
		Type:    ClientIPFromHeader,
		Headers: []string{"CF-Connecting-IP", "X-Real-IP"},
	})

	t.Run("first header present", func(t *testing.T) {
		got = ""
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("CF-Connecting-IP", "203.0.113.1")
		w := httptest.NewRecorder()
		mw(captureClientIP(t, &got)).ServeHTTP(w, req)
		if got != "203.0.113.1" {
			t.Errorf("expected 203.0.113.1, got %q", got)
		}
	})

	t.Run("only second header present", func(t *testing.T) {
		got = ""
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Real-IP", "203.0.113.2")
		w := httptest.NewRecorder()
		mw(captureClientIP(t, &got)).ServeHTTP(w, req)
		if got != "203.0.113.2" {
			t.Errorf("expected 203.0.113.2, got %q", got)
		}
	})

	t.Run("neither header present falls through to next", func(t *testing.T) {
		nextCalled := false
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(w, req)
		if !nextCalled {
			t.Error("expected next handler to be called when no matching header is present")
		}
	})
}

func TestClientIP_XForwardedFor(t *testing.T) {
	var got string
	mw := ClientIP(&ClientIPConfig{
		Type:              ClientIPFromXForwardedFor,
		TrustedIPPrefixes: []string{"10.0.0.0/8"},
	})

	t.Run("client before trusted proxy", func(t *testing.T) {
		got = ""
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		// Chain: client → trusted proxy; rightmost trusted is skipped.
		req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
		w := httptest.NewRecorder()
		mw(captureClientIP(t, &got)).ServeHTTP(w, req)
		if got != "203.0.113.5" {
			t.Errorf("expected 203.0.113.5, got %q", got)
		}
	})

	t.Run("untrusted IP is client directly", func(t *testing.T) {
		got = ""
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", "198.51.100.7")
		w := httptest.NewRecorder()
		mw(captureClientIP(t, &got)).ServeHTTP(w, req)
		if got != "198.51.100.7" {
			t.Errorf("expected 198.51.100.7, got %q", got)
		}
	})

	t.Run("all trusted leaves no client IP", func(t *testing.T) {
		got = ""
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.2, 10.0.0.1")
		w := httptest.NewRecorder()
		mw(captureClientIP(t, &got)).ServeHTTP(w, req)
		if got != "" {
			t.Errorf("expected empty client IP when all hops are trusted, got %q", got)
		}
	})
}

func TestClientIP_XForwardedForTrustedProxies(t *testing.T) {
	// chi's walkXFF traverses right-to-left and decrements n each step.
	// With numTrustedProxies=1, the walk stops at the 1st entry from the right
	// (n reaches 0), so that entry becomes the resolved client IP.
	// Chain "client, proxy" → walk hits "proxy" first (n→0, entry="proxy").
	// This reflects chi's design: use the N-th hop from the right as the boundary.
	mw := ClientIP(&ClientIPConfig{
		Type:              ClientIPFromXForwardForTrustedProxies,
		NumTrustedProxies: 1,
	})
	var got string

	t.Run("selects rightmost entry as client IP with one trusted proxy", func(t *testing.T) {
		got = ""
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.10, 10.0.0.1")
		w := httptest.NewRecorder()
		mw(captureClientIP(t, &got)).ServeHTTP(w, req)
		// rightmost = "10.0.0.1" (first visited, n→0, entry set)
		if got != "10.0.0.1" {
			t.Errorf("expected 10.0.0.1, got %q", got)
		}
	})

	t.Run("two trusted proxies skips rightmost, selects second from right", func(t *testing.T) {
		got = ""
		mw2 := ClientIP(&ClientIPConfig{
			Type:              ClientIPFromXForwardForTrustedProxies,
			NumTrustedProxies: 2,
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.10, 10.0.0.2, 10.0.0.1")
		w := httptest.NewRecorder()
		mw2(captureClientIP(t, &got)).ServeHTTP(w, req)
		// walk: "10.0.0.1" (n→1), "10.0.0.2" (n→0, entry set)
		if got != "10.0.0.2" {
			t.Errorf("expected 10.0.0.2, got %q", got)
		}
	})

	t.Run("fewer entries than trusted proxies leaves no client IP", func(t *testing.T) {
		got = ""
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		// Chain has 0 entries — n never reaches 0, no entry is set.
		req.Header.Set("X-Forwarded-For", "")
		w := httptest.NewRecorder()
		mw(captureClientIP(t, &got)).ServeHTTP(w, req)
		if got != "" {
			t.Errorf("expected empty client IP, got %q", got)
		}
	})
}
