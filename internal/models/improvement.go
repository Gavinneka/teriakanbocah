package models

import "time"

type ImprovementLog struct {
	ID        int
	Title     string
	Cost      int    // dalam Rupiah
	Status    string // selesai, proses, ditunda
	Notes     string
	DoneAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
