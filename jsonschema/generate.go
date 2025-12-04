package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/relychan/gohttps"
	"github.com/relychan/goutils"
)

func main() {
	err := jsonSchemaConfiguration()
	if err != nil {
		panic(fmt.Errorf("failed to write jsonschema for ServerConfig: %w", err))
	}
}

func jsonSchemaConfiguration() error {
	r := new(jsonschema.Reflector)

	err := r.AddGoComments(
		"github.com/relychan/gohttps",
		"..",
		jsonschema.WithFullComment(),
	)
	if err != nil {
		return err
	}

	reflectSchema := r.Reflect(gohttps.ServerConfig{})

	// custom schema types
	reflectSchema.Definitions["Duration"] = &jsonschema.Schema{
		Type:        "string",
		Description: "Duration string",
		MinLength:   goutils.ToPtr(uint64(2)),
		Pattern:     `^(-?\d+(\.\d+)?h)?(-?\d+(\.\d+)?m)?(-?\d+(\.\d+)?s)?(-?\d+(\.\d+)?ms)?$`,
	}

	reflectSchema.Definitions["ServerConfig"].Properties.Set("requestTimeout", &jsonschema.Schema{
		Description: "The default timeout of every request. Return a 504 Gateway Timeout error to the client.",
		Ref:         "#/$defs/Duration",
	})
	reflectSchema.Definitions["ServerConfig"].Properties.Set("readTimeout", &jsonschema.Schema{
		Description: "The maximum duration for reading the entire request, including the body.\nA zero or negative value means there will be no timeout.",
		Ref:         "#/$defs/Duration",
	})
	reflectSchema.Definitions["ServerConfig"].Properties.Set(
		"readHeaderTimeout",
		&jsonschema.Schema{
			Description: `The amount of time allowed to read request headers.
The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body.
If zero, the value of ReadTimeout is used. If negative, or if zero and ReadTimeout is zero or negative, there is no timeout.`,
			Ref: "#/$defs/Duration",
		},
	)
	reflectSchema.Definitions["ServerConfig"].Properties.Set("writeTimeout", &jsonschema.Schema{
		Description: "The maximum duration before timing out writes of the response. It is reset whenever a new request's header is read.\nLike ReadTimeout, it does not let Handlers make decisions on a per-request basis.\nA zero or negative value means there will be no timeout.",
		Ref:         "#/$defs/Duration",
	})
	reflectSchema.Definitions["ServerConfig"].Properties.Set("idleTimeout", &jsonschema.Schema{
		Description: "The maximum amount of time to wait for the next request when keep-alives are enabled.\nIf zero, the value of ReadTimeout is used.\nIf negative, or if zero and ReadTimeout is zero or negative, there is no timeout.",
		Ref:         "#/$defs/Duration",
	})

	buffer := new(bytes.Buffer)
	enc := json.NewEncoder(buffer)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")

	err = enc.Encode(reflectSchema)
	if err != nil {
		return err
	}

	return os.WriteFile( //nolint:gosec
		filepath.Join("jsonschema", "server.schema.json"),
		buffer.Bytes(), 0o644,
	)
}
