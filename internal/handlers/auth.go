package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"
	"weight-tracker/internal/models"
)

type AuthHandler struct {
	userRepo *models.UserRepository
	tmpl     *template.Template
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/partials/*.html"))

	return &AuthHandler{
		userRepo: models.NewUserRepository(db),
		tmpl:     tmpl,
	}
}

func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := map[string]interface{}{
			"Registered": r.URL.Query().Get("registered") == "true",
		}
		h.tmpl.ExecuteTemplate(w, "login.html", data)
		return
	}

	// Handle POST - login form submission
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		h.tmpl.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"Error": "Username and password are required",
		})
		return
	}

	user, err := h.userRepo.GetByUsername(username)
	if err != nil || !h.userRepo.VerifyPassword(user, password) {
		h.tmpl.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"Error": "Invalid username or password",
		})
		return
	}

	// Set session cookie (simplified - in production use secure session management)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "user_" + username, // Simplified - use proper session tokens
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) ShowRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.tmpl.ExecuteTemplate(w, "register.html", nil)
		return
	}

	// Handle POST - registration form submission
	username := r.FormValue("username")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if username == "" || password == "" {
		h.tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"Error": "Username and password are required",
		})
		return
	}

	if password != confirmPassword {
		h.tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"Error": "Passwords do not match",
		})
		return
	}

	if len(password) < 6 {
		h.tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"Error": "Password must be at least 6 characters",
		})
		return
	}

	// Check if user already exists
	_, err := h.userRepo.GetByUsername(username)
	if err == nil {
		h.tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"Error": "Username already exists",
		})
		return
	}

	// Create new user
	_, err = h.userRepo.Create(username, password)
	if err != nil {
		log.Printf("Failed to create user %s: %v", username, err)
		errorMsg := "Failed to create user"
		// Provide more specific error messages for debugging
		if err.Error() == "UNIQUE constraint failed" {
			errorMsg = "Username already exists"
		} else if err.Error() == "database is locked" {
			errorMsg = "Database busy, please try again"
		}
		h.tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"Error": errorMsg,
		})
		return
	}

	// Redirect to login
	http.Redirect(w, r, "/login?registered=true", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}