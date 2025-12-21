package middleware

import (
	"context"
	"net/http"
	"strings"

	authv1 "gateway/auth/v1"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func NewJWT(client authv1.AuthServiceClient) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]

			resp, err := client.Validate(
				r.Context(),
				&authv1.ValidateRequest{
					Token: tokenStr,
				},
			)

			if err != nil || !resp.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, resp.UserId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok
}
