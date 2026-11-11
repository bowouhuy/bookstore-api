package di

import (
	"errors"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"

	"github.com/bowouhuy/bookstore-api/config"
)

var (
	cfg     *config.Config
	cfgOnce sync.Once
)

// GetConfig loads .env once; variables already set in the environment take precedence.
func GetConfig() *config.Config {
	cfgOnce.Do(func() {
		if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Fatalf("load .env: %v", err)
		}

		cfg = config.New()
		if err := cfg.Validate(); err != nil {
			log.Fatalf("invalid config: %v", err)
		}
	})

	return cfg
}
