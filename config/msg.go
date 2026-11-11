package config

import _ "embed"

// Messages maps HTTP status codes to response messages per language.
// It is embedded so the binary works regardless of the working directory.
//
//go:embed msg.yaml
var Messages []byte
