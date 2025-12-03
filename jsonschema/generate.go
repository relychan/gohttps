package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/relychan/gohttps"
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
