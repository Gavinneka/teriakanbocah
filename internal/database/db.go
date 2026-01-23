package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite", "./ac_maintenance.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS maintenance_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room TEXT,
		unit TEXT,
		activity TEXT,
		date DATETIME,
		status TEXT,
		notes TEXT
	);
	
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password TEXT,
		role TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS permissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		app_name TEXT,
		capabilities TEXT,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		outcome_goal TEXT,
		status TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		outcome TEXT,
		estimate INTEGER,
		status TEXT DEFAULT 'inbox',
		project_id INTEGER,
		scheduled_date DATE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		FOREIGN KEY(project_id) REFERENCES projects(id)
	);

	CREATE TABLE IF NOT EXISTS problem_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		problem TEXT,
		cause TEXT,
		solution TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS reviews (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		week_start DATE,
		review_notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS task_obstacles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER,
		description TEXT,
		status TEXT DEFAULT 'open', -- open, resolved
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		resolved_at DATETIME,
		FOREIGN KEY(task_id) REFERENCES tasks(id)
	);`

	_, err = DB.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Hotfix: Ensure completed_at exists for existing databases
	_, _ = DB.Exec("ALTER TABLE tasks ADD COLUMN completed_at DATETIME;")
	// Hotfix: Add priority and is_archived
	_, _ = DB.Exec("ALTER TABLE tasks ADD COLUMN priority TEXT DEFAULT 'medium';")
	_, _ = DB.Exec("ALTER TABLE tasks ADD COLUMN is_archived BOOLEAN DEFAULT 0;")

	// User Management Enhancements
	_, _ = DB.Exec("ALTER TABLE users ADD COLUMN last_login DATETIME;")
	_, _ = DB.Exec("ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT 1;")
	_, _ = DB.Exec("ALTER TABLE users ADD COLUMN allowed_modules TEXT DEFAULT 'ac,work';") // Comma-separated: ac,work,admin

	// Login Logs Table
	_, _ = DB.Exec(`CREATE TABLE IF NOT EXISTS login_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		username TEXT,
		login_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		ip_address TEXT,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`)
}
