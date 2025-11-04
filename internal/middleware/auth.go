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

			// For simplicity, we'll just check if a session exists
			// In a real app, you'd validate the session token
			// and retrieve the user from the database
			ctx := context.WithValue(r.Context(), IsAuthKey, true)

			// TODO: Implement proper session validation and user retrieval
			// userID := validateSessionToken(cookie.Value)
			// if userID != 0 {
			//     user, err := userRepo.GetByID(userID)
			//     if err == nil {
			//         ctx = context.WithValue(ctx, UserIDKey, user.ID)
			//         ctx = context.WithValue(ctx, UserKey, user)
			//         ctx = context.WithValue(ctx, IsAuthKey, true)
			//     }
			// }

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuthenticated := r.Context().Value(IsAuthKey).(bool)
		if !isAuthenticated {
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