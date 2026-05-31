package main

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
	"github.com/relychan/gohttps/middlewares"
)

type ClientIPConfig middlewares.ClientIPConfig

// JSONSchema defines a custom definition for JSON schema.
func (ClientIPConfig) JSONSchema() *jsonschema.Schema {
	keyType := "type"
	keyHeaders := "headers"
	keyTrustedIPPrefixes := "trustedIpPrefixes"
	keyNumTrustedProxies := "numTrustedProxies"
	typeDescription := "Type of the strategy that the client IP should be parsed from."

	xffProps := jsonschema.NewProperties()
	xffProps.Set(keyType, &jsonschema.Schema{
		Type:        "string",
		Description: typeDescription,
		Enum:        []any{middlewares.ClientIPFromXForwardedFor},
	})
	xffProps.Set(keyTrustedIPPrefixes, &jsonschema.Schema{
		Type:        "array",
		Description: "List of CIDR prefixes to be trusted when parsing the client IP from the X-Forwarded-For header.",
		MinItems:    new(uint64(1)),
		Items: &jsonschema.Schema{
			Type:    "string",
			Pattern: `((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\/([0-9]+))|((([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))\/([0-9]+))`,
		},
	})

	xffTrustedProxiesProps := jsonschema.NewProperties()
	xffTrustedProxiesProps.Set(keyType, &jsonschema.Schema{
		Type:        "string",
		Description: typeDescription,
		Enum:        []any{middlewares.ClientIPFromXForwardForTrustedProxies},
	})
	xffTrustedProxiesProps.Set(keyNumTrustedProxies, &jsonschema.Schema{
		Type:        "integer",
		Description: "The exact number of trusted reverse proxies between this server and the public internet.",
		Minimum:     json.Number("1"),
		Default:     1,
	})

	headerProps := jsonschema.NewProperties()
	headerProps.Set(keyType, &jsonschema.Schema{
		Type:        "string",
		Description: typeDescription,
		Enum:        []any{middlewares.ClientIPFromHeader},
	})
	headerProps.Set(keyHeaders, &jsonschema.Schema{
		Type:        "array",
		Description: "List of headers to be looked up",
		MinItems:    new(uint64(1)),
		Items: &jsonschema.Schema{
			Type:      "string",
			MinLength: new(uint64(1)),
		},
	})

	remoteAddrProps := jsonschema.NewProperties()
	remoteAddrProps.Set(keyType, &jsonschema.Schema{
		Type:        "string",
		Description: typeDescription,
		Enum:        []any{middlewares.ClientIPFromRemoteAddr},
	})

	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Title:       "ServerClientIPFromHeaderConfig",
				Description: "Configuration for client IP resolution from headers. Only safe with headers your proxy unconditionally OVERWRITES on every request.",
				Type:        "object",
				Properties:  headerProps,
				Required:    []string{keyType, keyHeaders},
			},
			{
				Title:       "ServerClientIPFromRemoteAddressConfig",
				Description: "Configuration for client IP resolution from the remote address of the incoming request — the IP address of whoever opened the connection to this server. Use this strategy when this server is directly connected to the public internet with NO reverse proxy in front of it. Behind a reverse proxy, RemoteAddr is the proxy's IP, not the client's — use ClientIPFromHeader or ClientIPFromXFF instead",
				Type:        "object",
				Properties:  remoteAddrProps,
				Required:    []string{keyType},
			},
			{
				Title:       "ServerClientIPFromXForwardForConfig",
				Description: "Configuration for client IP resolution from X-Forward-For header with trusted IP prefixes, walking the chain right-to-left and skipping any IP that falls within one of the given trusted CIDR prefixes.",
				Type:        "object",
				Properties:  xffProps,
				Required:    []string{keyType, keyTrustedIPPrefixes},
			},
			{
				Title:       "ServerClientIPFromXForwardForTrustedProxiesConfig",
				Description: "Configuration for client IP resolution from X-Forward-For header given the exact number of trusted reverse proxies between this server and the public internet. It returns the IP at position len(xff) - numTrustedProxies in the merged X-Forwarded-For list — the IP added by the outermost of your trusted proxies, the only IP in the chain that none of your proxies have allowed an attacker to forge.",
				Type:        "object",
				Properties:  xffTrustedProxiesProps,
				Required:    []string{keyType, keyNumTrustedProxies},
			},
		},
	}
}
