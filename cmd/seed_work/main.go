package main

import (
	"ac_tracker/internal/database"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	database.InitDB()
	db := database.DB

	// Seed Projects
	projects := []struct {
		Name        string
		OutcomeGoal string
		Status      string
	}{
		{"Upgrade Server", "Migrate to new infrastructure", "active"},
		{"Kantor Baru", "Persiapan pindahan kantor", "active"},
		{"Marketing Q1", "Target reach 1M users", "planned"},
	}

	for _, p := range projects {
		var id int
		err := db.QueryRow("SELECT id FROM projects WHERE name = ?", p.Name).Scan(&id)
		if err == sql.ErrNoRows {
			res, err := db.Exec("INSERT INTO projects(name, outcome_goal, status, created_at) VALUES(?, ?, ?, ?)",
				p.Name, p.OutcomeGoal, p.Status, time.Now())
			if err != nil {
				log.Printf("Error creating project %s: %v", p.Name, err)
			} else {
				fmt.Printf("Created project: %s\n", p.Name)
				lid, _ := res.LastInsertId()
				id = int(lid)
			}
		} else {
			fmt.Printf("Project %s exists.\n", p.Name)
		}

		// Seed Tasks for this project
		if p.Name == "Upgrade Server" {
			seedTask(db, "Beli VPS Baru", "VPS terbeli", "todo", id)
			seedTask(db, "Setup Environment", "Env siap", "todo", id)
		} else if p.Name == "Kantor Baru" {
			seedTask(db, "Cari Lokasi", "Dapat lokasi strategis", "done", id)
			seedTask(db, "Beli Furniture", "Meja kursi lengkap", "doing", id)
		}
	}

	// Seed Loose Tasks (Inbox)
	seedTask(db, "Cek Email Client", "Client terbalas", "inbox", 0)
	seedTask(db, "Bayar Tagihan Internet", "Internet aman", "todo", 0)
}

func seedTask(db *sql.DB, title, outcome, status string, projectID int) {
	var exists int
	err := db.QueryRow("SELECT COUNT(*) FROM tasks WHERE title = ?", title).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return
	}

	if exists == 0 {
		var res sql.Result
		var err error
		if projectID > 0 {
			res, err = db.Exec("INSERT INTO tasks(title, outcome, status, project_id, created_at, updated_at, priority) VALUES(?, ?, ?, ?, ?, ?, ?)",
				title, outcome, status, projectID, time.Now(), time.Now(), "medium")
		} else {
			res, err = db.Exec("INSERT INTO tasks(title, outcome, status, created_at, updated_at, priority) VALUES(?, ?, ?, ?, ?, ?)",
				title, outcome, status, time.Now(), time.Now(), "medium")
		}

		if err != nil {
			log.Printf("Error seeding task %s: %v", title, err)
		} else {
			fmt.Printf("Seeded task: %s\n", title)
			// Add obstacle to one specific task for demo
			if title == "Beli Furniture" {
				lid, _ := res.LastInsertId()
				db.Exec("INSERT INTO task_obstacles(task_id, description, status, created_at) VALUES(?, ?, ?, ?)",
					lid, "Budget belum cair", "open", time.Now())
				fmt.Println("  -> Added obstacle: Budget belum cair")
			}
		}
	} else {
		fmt.Printf("Task %s exists.\n", title)
	}
}
