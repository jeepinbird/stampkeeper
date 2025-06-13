package middleware

import (
	"context"
)

// Context keys for storing preferences in request context
type contextKey string

const preferencesKey contextKey = "user_preferences"

// WithPreferences adds user preferences to the context
func WithPreferences(ctx context.Context, prefs UserPreferences) context.Context {
	return context.WithValue(ctx, preferencesKey, prefs)
}

// GetPreferencesFromContext retrieves user preferences from the context
func GetPreferencesFromContext(ctx context.Context) (UserPreferences, bool) {
	prefs, ok := ctx.Value(preferencesKey).(UserPreferences)
	if !ok {
		return DefaultPreferences(), false
	}
	return prefs, true
}

// MustGetPreferencesFromContext retrieves preferences from context or returns defaults
func MustGetPreferencesFromContext(ctx context.Context) UserPreferences {
	prefs, ok := GetPreferencesFromContext(ctx)
	if !ok {
		return DefaultPreferences()
	}
	return prefs
}