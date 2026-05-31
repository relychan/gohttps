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

	xffProps := jsonschema.NewProperties()
	xffProps.Set(keyType, &jsonschema.Schema{
		Type: "string",
		Enum: []any{middlewares.ClientIPFromXForwardedFor},
	})
	xffProps.Set(keyTrustedIPPrefixes, &jsonschema.Schema{
		Type:     "array",
		MinItems: new(uint64(1)),
		Items: &jsonschema.Schema{
			Type:    "string",
			Pattern: `((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\/([0-9]+))|((([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))\/([0-9]+))`,
		},
	})

	xffTrustedProxiesProps := jsonschema.NewProperties()
	xffTrustedProxiesProps.Set(keyType, &jsonschema.Schema{
		Type: "string",
		Enum: []any{middlewares.ClientIPFromXForwardForTrustedProxies},
	})
	xffTrustedProxiesProps.Set(keyNumTrustedProxies, &jsonschema.Schema{
		Type:    "integer",
		Minimum: json.Number("1"),
		Default: 1,
	})

	headerProps := jsonschema.NewProperties()
	headerProps.Set(keyType, &jsonschema.Schema{
		Type: "string",
		Enum: []any{middlewares.ClientIPFromHeader},
	})
	headerProps.Set(keyHeaders, &jsonschema.Schema{
		Type:     "array",
		MinItems: new(uint64(1)),
		Items: &jsonschema.Schema{
			Type:      "string",
			MinLength: new(uint64(1)),
		},
	})

	remoteAddrProps := jsonschema.NewProperties()
	remoteAddrProps.Set(keyType, &jsonschema.Schema{
		Type: "string",
		Enum: []any{middlewares.ClientIPFromRemoteAddr},
	})

	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Title:       "ServerClientIPFromHeaderConfig",
				Description: "Configuration for client IP resolution from header",
				Type:        "object",
				Properties:  headerProps,
				Required:    []string{keyType, keyHeaders},
			},
			{
				Title:       "ServerClientIPFromRemoteAddressConfig",
				Description: "Configuration for client IP resolution from remote address",
				Type:        "object",
				Properties:  remoteAddrProps,
				Required:    []string{keyType},
			},
			{
				Title:       "ServerClientIPFromXForwardForConfig",
				Description: "Configuration for client IP resolution from X-Forward-For headers with trusted IP prefixes",
				Type:        "object",
				Properties:  xffProps,
				Required:    []string{keyType, keyTrustedIPPrefixes},
			},
			{
				Title:       "ServerClientIPFromXForwardForTrustedProxiesConfig",
				Description: "Configuration for client IP resolution from X-Forward-For headers with trusted proxies",
				Type:        "object",
				Properties:  xffTrustedProxiesProps,
				Required:    []string{keyType, keyNumTrustedProxies},
			},
		},
	}
}
