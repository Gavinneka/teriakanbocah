package models

import "time"

type MaintenanceRecord struct {
	ID        int       `json:"id"`
	Room      string    `json:"room"`
	Unit      string    `json:"unit"`
	Activity  string    `json:"activity"`
	Date      time.Time `json:"date"`
	Status    string    `json:"status"` // Scheduled, Completed, Pending
	Notes     string    `json:"notes"`
}
