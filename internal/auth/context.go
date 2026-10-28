package auth

import "context"

// contextKey is unexported so keys from other packages cannot collide with it.
type contextKey struct{}

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, contextKey{}, userID)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(contextKey{}).(int64)
	return userID, ok
}
