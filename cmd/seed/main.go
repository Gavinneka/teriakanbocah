package main

import (
	"ac_tracker/internal/database"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type SeedData struct {
	ID       int
	Room     string
	Unit     string
	Activity string
	Date     string
	Status   string
	Notes    string
}

func main() {
	database.InitDB()
	db := database.DB

	data := []SeedData{
		{1, "Lab", "Samsung x2", "", "", "", ""},
		{2, "FO", "Daikin x2", "", "", "", ""},
		{3, "Utama", "Samsung, sharp", "Cuci Filter", "09 Desember 2025", "", ""},
		{4, "1.1 (Gudang)", "Samsung x2", "", "", "", "Kanan kompresor rusak"},
		{5, "1.2 (TTL)", "Samsung x2", "", "", "", ""},
		{6, "1.3 (Tempat Motor)", "Samsung x2", "", "", "", "Kanan kompresor rusak"},
		{7, "2.1 (Ruang kelas)", "Daikin", "Baru pasang", "26 September 2025", "", ""},
		{8, "2.2 (Ruang kelas)", "Daikin x2", "Baru pasang", "26 September 2025", "", ""},
		{9, "2.3 (Ruang kelas)", "Daikin x2", "Baru pasang", "19 September 2025", "", ""},
		{10, "2.4 (Ruang kelas)", "Daikin x2", "Baru pasang", "19 September 2025", "", ""},
		{11, "2.5 (Ruang kelas)", "Samsung, panasonic", "Cuci Filter", "10 Desember 2025", "", "Panasonic controlnya error"},
		{12, "2.6 (Ruang kelas)", "Samsung, LG x2, Panasonic", "", "", "", ""},
		{13, "3.1 (Aula)", "Daikin 1pk, 2pk, 1pk, 2pk", "Baru pasang", "5 September 2025", "", ""},
	}

	monthMap := map[string]string{
		"Januari":   "01",
		"Februari":  "02",
		"Maret":     "03",
		"April":     "04",
		"Mei":       "05",
		"Juni":      "06",
		"Juli":      "07",
		"Agustus":   "08",
		"September": "09",
		"Oktober":   "10",
		"November":  "11",
		"Desember":  "12",
	}

	for _, d := range data {
		// Check if exists
		var exists int
		err := db.QueryRow("SELECT COUNT(*) FROM maintenance_records WHERE room = ? AND unit = ? AND date = ?", d.Room, d.Unit, parseDate(d.Date, monthMap)).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			log.Println("Error checking existence:", err)
		}
		if exists > 0 {
			fmt.Printf("Skipping existing: %s\n", d.Room)
			continue
		}

		parsedDate := parseDate(d.Date, monthMap)
		status := d.Status
		if status == "" {
			if d.Activity == "Baru pasang" || d.Activity == "Cuci Filter" {
				status = "Selesai"
			} else {
				status = "Data Awal"
			}
		}

		_, err = db.Exec("INSERT INTO maintenance_records(room, unit, activity, date, status, notes) VALUES(?, ?, ?, ?, ?, ?)",
			d.Room, d.Unit, d.Activity, parsedDate, status, d.Notes)
		if err != nil {
			log.Printf("Error inserting %s: %v\n", d.Room, err)
		} else {
			fmt.Printf("Inserted: %s\n", d.Room)
		}
	}
}

func parseDate(dateStr string, monthMap map[string]string) time.Time {
	if dateStr == "" {
		// Default to today or specific past date for initial data?
		// Let's use a fixed date for consistency if empty, or just today.
		// Given it's "inventory", maybe 2024-01-01? Or current time.
		return time.Now()
	}

	parts := strings.Split(dateStr, " ")
	if len(parts) >= 3 {
		day := parts[0]
		if len(day) == 1 {
			day = "0" + day
		}
		monthName := parts[1]
		year := parts[2]
		
		monthNum, ok := monthMap[monthName]
		if !ok {
			return time.Now()
		}

		// RFC3339-like yyyy-mm-dd
		t, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", year, monthNum, day))
		if err != nil {
			return time.Now()
		}
		return t
	}
	return time.Now()
}
