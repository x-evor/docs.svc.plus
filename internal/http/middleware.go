package httpapi

import "net/http"

func RequireServiceToken(expected string, next http.Handler) http.Handler {
	if expected == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Service-Token") != expected {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"error": "invalid_service_token",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
