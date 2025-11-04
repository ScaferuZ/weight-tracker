package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database"`
	Version   string    `json:"version"`
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	var dbStatus string
	if err := h.db.Ping(); err != nil {
		dbStatus = "error"
	} else {
		dbStatus = "healthy"
	}

	// Determine overall health
	status := "healthy"
	if dbStatus != "healthy" {
		status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Database:  dbStatus,
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}