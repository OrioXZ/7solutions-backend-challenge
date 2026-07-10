package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/security"
)

type authContextKey string

const authenticatedSubjectKey authContextKey = "authenticated-subject"

// AuthenticationMiddleware validates Bearer tokens before protected handlers run.
func AuthenticationMiddleware(validator security.TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := strings.TrimSpace(r.Header.Get("Authorization"))
			parts := strings.Fields(authorization)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "a valid bearer token is required")
				return
			}

			claims, err := validator.Validate(parts[1])
			if err != nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "a valid bearer token is required")
				return
			}

			ctx := context.WithValue(r.Context(), authenticatedSubjectKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthenticatedUserID returns the JWT subject stored by AuthenticationMiddleware.
func AuthenticatedUserID(ctx context.Context) (string, bool) {
	subject, ok := ctx.Value(authenticatedSubjectKey).(string)
	return subject, ok && subject != ""
}
