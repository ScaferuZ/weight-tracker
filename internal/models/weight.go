package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Weight struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	WeightKg   float64   `json:"weight_kg"`
	RecordedAt time.Time `json:"recorded_at"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type WeightRepository struct {
	db *sql.DB
}

func NewWeightRepository(db *sql.DB) *WeightRepository {
	return &WeightRepository{db: db}
}

func (r *WeightRepository) Create(weight *Weight) error {
	query := `INSERT INTO weights (user_id, weight_kg, recorded_at, notes) VALUES (?, ?, ?, ?)`
	result, err := r.db.Exec(query, weight.UserID, weight.WeightKg, weight.RecordedAt, weight.Notes)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	weight.ID = int(id)
	return nil
}

func (r *WeightRepository) GetByDate(userID int, date string) (*Weight, error) {
	query := `SELECT id, user_id, weight_kg, recorded_at, notes, created_at, updated_at
              FROM weights WHERE user_id = ? AND DATE(recorded_at) = DATE(?)`
	row := r.db.QueryRow(query, userID, date)

	var weight Weight
	err := row.Scan(&weight.ID, &weight.UserID, &weight.WeightKg, &weight.RecordedAt,
		&weight.Notes, &weight.CreatedAt, &weight.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &weight, nil
}

func (r *WeightRepository) Update(weight *Weight) error {
	query := `UPDATE weights SET weight_kg = ?, recorded_at = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
              WHERE id = ? AND user_id = ?`
	_, err := r.db.Exec(query, weight.WeightKg, weight.RecordedAt, weight.Notes, weight.ID, weight.UserID)
	return err
}

func (r *WeightRepository) GetRecent(userID int, limit int) ([]Weight, error) {
	query := `SELECT id, user_id, weight_kg, recorded_at, notes, created_at, updated_at
              FROM weights WHERE user_id = ? ORDER BY recorded_at DESC LIMIT ?`
	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weights []Weight
	for rows.Next() {
		var weight Weight
		err := rows.Scan(&weight.ID, &weight.UserID, &weight.WeightKg, &weight.RecordedAt,
			&weight.Notes, &weight.CreatedAt, &weight.UpdatedAt)
		if err != nil {
			return nil, err
		}
		weights = append(weights, weight)
	}

	return weights, nil
}

func (r *WeightRepository) Delete(id, userID int) error {
	query := `DELETE FROM weights WHERE id = ? AND user_id = ?`
	_, err := r.db.Exec(query, id, userID)
	return err
}

func (r *WeightRepository) GetChartData(userID int, days int) ([]Weight, error) {
	query := `SELECT id, user_id, weight_kg, recorded_at, notes, created_at, updated_at
              FROM weights
              WHERE user_id = ? AND recorded_at >= date('now', '-{} days')
              ORDER BY recorded_at ASC`

	// Use string formatting for the days parameter since it's a number
	rows, err := r.db.Query(fmt.Sprintf(query, days), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weights []Weight
	for rows.Next() {
		var weight Weight
		err := rows.Scan(&weight.ID, &weight.UserID, &weight.WeightKg, &weight.RecordedAt,
			&weight.Notes, &weight.CreatedAt, &weight.UpdatedAt)
		if err != nil {
			return nil, err
		}
		weights = append(weights, weight)
	}

	return weights, nil
}