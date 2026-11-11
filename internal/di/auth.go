package di

import (
	"sync"

	"github.com/bowouhuy/bookstore-api/internal/middleware"
	"github.com/bowouhuy/bookstore-api/pkg/auth"
)

var (
	tokenManager     *auth.TokenManager
	tokenManagerOnce sync.Once
)

func GetTokenManager() *auth.TokenManager {
	tokenManagerOnce.Do(func() {
		cfg := GetConfig()
		tokenManager = auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.TTL)
	})

	return tokenManager
}

func GetAuthMiddleware() middleware.Handle {
	return middleware.Auth(GetTokenManager())
}
