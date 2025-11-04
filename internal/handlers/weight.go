package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"
	"weight-tracker/internal/middleware"
	"weight-tracker/internal/models"
)

type WeightHandler struct {
	weightRepo *models.WeightRepository
	tmpl       *template.Template
}

func NewWeightHandler(db *sql.DB) *WeightHandler {
	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/partials/*.html"))

	return &WeightHandler{
		weightRepo: models.NewWeightRepository(db),
		tmpl:       tmpl,
	}
}

func (h *WeightHandler) ShowWeights(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	weights, err := h.weightRepo.GetRecent(userID, 50)
	if err != nil {
		http.Error(w, "Failed to load weights", http.StatusInternalServerError)
		return
	}

	// Check if user has entry for today
	today := time.Now().Format("2006-01-02")
	todayWeight, _ := h.weightRepo.GetByDate(userID, today)

	data := map[string]interface{}{
		"Weights":      weights,
		"HasTodayEntry": todayWeight != nil,
		"TodayWeight":  todayWeight,
	}

	h.tmpl.ExecuteTemplate(w, "weights.html", data)
}

func (h *WeightHandler) CreateWeight(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	weightStr := r.FormValue("weight")
	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil || weight <= 0 {
		http.Error(w, "Invalid weight value", http.StatusBadRequest)
		return
	}

	if weight < 20 || weight > 500 {
		http.Error(w, "Weight must be between 20 and 500 kg", http.StatusBadRequest)
		return
	}

	notes := r.FormValue("notes")
	today := time.Now().Format("2006-01-02")

	// Check if entry exists for today
	existingWeight, err := h.weightRepo.GetByDate(userID, today)

	if err == nil && existingWeight != nil {
		// Update existing entry
		existingWeight.WeightKg = weight
		existingWeight.Notes = notes
		if err := h.weightRepo.Update(existingWeight); err != nil {
			http.Error(w, "Failed to update weight", http.StatusInternalServerError)
			return
		}
	} else {
		// Create new entry
		newWeight := &models.Weight{
			UserID:     userID,
			WeightKg:   weight,
			RecordedAt: time.Now(),
			Notes:      notes,
		}
		if err := h.weightRepo.Create(newWeight); err != nil {
			http.Error(w, "Failed to create weight", http.StatusInternalServerError)
			return
		}
	}

	// If HTMX request, return partial template
	if r.Header.Get("HX-Request") == "true" {
		weights, err := h.weightRepo.GetRecent(userID, 10)
		if err != nil {
			http.Error(w, "Failed to load weights", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Weights": weights,
		}

		h.tmpl.ExecuteTemplate(w, "weight_list.html", data)
		return
	}

	// Regular form submission - redirect
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *WeightHandler) DeleteWeight(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	weightIDStr := r.URL.Query().Get("id")
	weightID, err := strconv.Atoi(weightIDStr)
	if err != nil {
		http.Error(w, "Invalid weight ID", http.StatusBadRequest)
		return
	}

	if err := h.weightRepo.Delete(weightID, userID); err != nil {
		http.Error(w, "Failed to delete weight", http.StatusInternalServerError)
		return
	}

	// If HTMX request, return empty response
	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}