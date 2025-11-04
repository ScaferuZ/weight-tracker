package middleware

import (
	"context"
	"net/http"
	"weight-tracker/internal/models"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserKey     contextKey = "user"
	IsAuthKey   contextKey = "is_authenticated"
)

func AuthMiddleware(userRepo *models.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for session cookie or token
			cookie, err := r.Cookie("session_token")
			if err != nil || cookie.Value == "" {
				// User not authenticated
				ctx := context.WithValue(r.Context(), IsAuthKey, false)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Validate session token format (should be "user_" + username)
			if len(cookie.Value) < 6 || cookie.Value[:5] != "user_" {
				// Invalid session format
				ctx := context.WithValue(r.Context(), IsAuthKey, false)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Extract username from session token
			username := cookie.Value[5:]
			if username == "" {
				// Invalid session - no username
				ctx := context.WithValue(r.Context(), IsAuthKey, false)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Verify user exists in database
			user, err := userRepo.GetByUsername(username)
			if err != nil {
				// User not found - invalid session
				ctx := context.WithValue(r.Context(), IsAuthKey, false)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// User is authenticated - set context values
			ctx := context.WithValue(r.Context(), IsAuthKey, true)
			ctx = context.WithValue(ctx, UserIDKey, user.ID)
			ctx = context.WithValue(ctx, UserKey, user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if user is authenticated
		authValue := r.Context().Value(IsAuthKey)
		if authValue == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		isAuthenticated, ok := authValue.(bool)
		if !ok || !isAuthenticated {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GetUserID(r *http.Request) int {
	if userID, ok := r.Context().Value(UserIDKey).(int); ok {
		return userID
	}
	return 0
}

func GetUser(r *http.Request) *models.User {
	if user, ok := r.Context().Value(UserKey).(*models.User); ok {
		return user
	}
	return nil
}

func IsAuthenticated(r *http.Request) bool {
	if isAuth, ok := r.Context().Value(IsAuthKey).(bool); ok {
		return isAuth
	}
	return false
}