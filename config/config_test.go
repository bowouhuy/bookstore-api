package config

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bowouhuy/bookstore-api/pkg/msg"
)

func TestConfig_Validate(t *testing.T) {
	valid := func() *Config {
		return &Config{
			App: App{Port: 8080},
			JWT: JWT{Secret: strings.Repeat("s", 32), TTL: time.Hour},
		}
	}

	tests := []struct {
		name    string
		mutate  func(c *Config)
		wantErr string
	}{
		{name: "valid", mutate: func(c *Config) {}},
		{name: "missing secret", mutate: func(c *Config) { c.JWT.Secret = "" }, wantErr: "JWT_SECRET is required"},
		{name: "short secret", mutate: func(c *Config) { c.JWT.Secret = "short" }, wantErr: "at least 32"},
		{name: "invalid ttl", mutate: func(c *Config) { c.JWT.TTL = 0 }, wantErr: "JWT_TTL"},
		{name: "invalid port", mutate: func(c *Config) { c.App.Port = 0 }, wantErr: "APP_PORT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := valid()
			tt.mutate(c)

			err := c.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestGetEnvAsDuration_InvalidValueFailsValidation(t *testing.T) {
	t.Setenv("JWT_TTL", "one day")

	if got := getEnvAsDuration("JWT_TTL", time.Hour); got != 0 {
		t.Errorf("got %v, want 0 so Validate rejects it", got)
	}
}

func TestMessages_CoverAllResponseCodes(t *testing.T) {
	if err := msg.Load(Messages); err != nil {
		t.Fatalf("load messages: %v", err)
	}

	codes := []int{
		http.StatusOK, http.StatusCreated, http.StatusNoContent,
		http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound,
		http.StatusMethodNotAllowed, http.StatusConflict, http.StatusRequestEntityTooLarge,
		http.StatusUnprocessableEntity, http.StatusInternalServerError, http.StatusServiceUnavailable,
	}

	for _, code := range codes {
		for _, lang := range []string{"en", "id"} {
			if msg.Get(code, lang) == "" {
				t.Errorf("missing %q message for code %d", lang, code)
			}
		}
	}
}
