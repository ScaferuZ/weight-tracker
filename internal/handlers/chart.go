package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
	"weight-tracker/internal/middleware"
	"weight-tracker/internal/models"
)

type ChartHandler struct {
	weightRepo *models.WeightRepository
}

func NewChartHandler(db *sql.DB) *ChartHandler {
	return &ChartHandler{
		weightRepo: models.NewWeightRepository(db),
	}
}

func (h *ChartHandler) GetWeightChartData(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	weights, err := h.weightRepo.GetChartData(userID, 90)
	if err != nil {
		http.Error(w, "Failed to fetch weight data", http.StatusInternalServerError)
		return
	}

	chartData := struct {
		Labels   []string `json:"labels"`
		Datasets []struct {
			Label           string    `json:"label"`
			Data            []float64 `json:"data"`
			BorderColor     string    `json:"borderColor"`
			BackgroundColor string    `json:"backgroundColor"`
			Fill            bool      `json:"fill"`
			Tension         float64   `json:"tension"`
		} `json:"datasets"`
	}{
		Labels:   []string{},
		Datasets: []struct {
			Label           string    `json:"label"`
			Data            []float64 `json:"data"`
			BorderColor     string    `json:"borderColor"`
			BackgroundColor string    `json:"backgroundColor"`
			Fill            bool      `json:"fill"`
			Tension         float64   `json:"tension"`
		}{
			{
				Label:           "Weight (kg)",
				Data:            []float64{},
				BorderColor:     "#3b82f6",
				BackgroundColor: "rgba(59, 130, 246, 0.1)",
				Fill:            true,
				Tension:         0.4,
			},
		},
	}

	for _, weight := range weights {
		chartData.Labels = append(chartData.Labels, weight.RecordedAt.Format("Jan 02"))
		chartData.Datasets[0].Data = append(chartData.Datasets[0].Data, weight.WeightKg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chartData)
}

func (h *ChartHandler) GetWeightStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	weights, err := h.weightRepo.GetRecent(userID, 1000) // Get more data for stats
	if err != nil {
		http.Error(w, "Failed to fetch weight data", http.StatusInternalServerError)
		return
	}

	stats := struct {
		CurrentWeight float64 `json:"current_weight"`
		Change7Days   float64 `json:"change_7_days"`
		Change30Days  float64 `json:"change_30_days"`
		AverageWeight float64 `json:"average_weight"`
		TotalEntries  int     `json:"total_entries"`
		MinWeight     float64 `json:"min_weight"`
		MaxWeight     float64 `json:"max_weight"`
	}{
		TotalEntries: len(weights),
	}

	if len(weights) > 0 {
		stats.CurrentWeight = weights[0].WeightKg
		stats.MinWeight = weights[0].WeightKg
		stats.MaxWeight = weights[0].WeightKg

		var totalWeight float64
		var weight7DaysAgo, weight30DaysAgo float64
		found7Days, found30Days := false, false

		for _, weight := range weights {
			totalWeight += weight.WeightKg

			if weight.WeightKg < stats.MinWeight {
				stats.MinWeight = weight.WeightKg
			}
			if weight.WeightKg > stats.MaxWeight {
				stats.MaxWeight = weight.WeightKg
			}

			// Find weight 7 days ago
			if !found7Days {
				daysDiff := time.Since(weight.RecordedAt).Hours() / 24
				if daysDiff >= 7 {
					weight7DaysAgo = weight.WeightKg
					found7Days = true
				}
			}

			// Find weight 30 days ago
			if !found30Days {
				daysDiff := time.Since(weight.RecordedAt).Hours() / 24
				if daysDiff >= 30 {
					weight30DaysAgo = weight.WeightKg
					found30Days = true
				}
			}
		}

		stats.AverageWeight = totalWeight / float64(len(weights))

		if found7Days {
			stats.Change7Days = stats.CurrentWeight - weight7DaysAgo
		}

		if found30Days {
			stats.Change30Days = stats.CurrentWeight - weight30DaysAgo
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}