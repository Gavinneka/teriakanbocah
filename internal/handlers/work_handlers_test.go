package handlers

import (
	"ac_tracker/internal/database"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB() {
	var err error
	database.DB, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}

	createTableSQL := `
	CREATE TABLE tasks (
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
		priority TEXT DEFAULT 'medium',
		is_archived BOOLEAN DEFAULT 0
	);
	CREATE TABLE task_obstacles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER,
		description TEXT,
		status TEXT DEFAULT 'open',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		resolved_at DATETIME
	);`

	_, err = database.DB.Exec(createTableSQL)
	if err != nil {
		panic(err)
	}
}

func TestCreateSimpleTask(t *testing.T) {
	setupTestDB()
	defer database.DB.Close()

	form := url.Values{}
	form.Add("title", "Test Task")
	req := httptest.NewRequest("POST", "/work/add", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	CreateSimpleTask(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", rr.Code)
	}

	var title string
	err := database.DB.QueryRow("SELECT title FROM tasks LIMIT 1").Scan(&title)
	if err != nil {
		t.Fatal(err)
	}
	if title != "Test Task" {
		t.Errorf("expected title 'Test Task', got '%s'", title)
	}
}

func TestAddObstacle(t *testing.T) {
	setupTestDB()
	defer database.DB.Close()

	// Insert a task first
	database.DB.Exec("INSERT INTO tasks (id, title) VALUES (1, 'Task with Obstacle')")

	form := url.Values{}
	form.Add("description", "Material shortage")
	req := httptest.NewRequest("POST", "/work/tasks/1/obstacles", strings.NewReader(form.Encode()))
	req.SetPathValue("id", "1")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	AddObstacle(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", rr.Code)
	}

	var description string
	err := database.DB.QueryRow("SELECT description FROM task_obstacles WHERE task_id = 1").Scan(&description)
	if err != nil {
		t.Fatal(err)
	}
	if description != "Material shortage" {
		t.Errorf("expected obstacle description 'Material shortage', got '%s'", description)
	}
}

func TestWorkDashboardFilteringQueries(t *testing.T) {
	setupTestDB()
	defer database.DB.Close()

	// Seed data
	database.DB.Exec("INSERT INTO tasks (id, title, status, is_archived) VALUES (1, 'Active Task', 'todo', 0)")
	database.DB.Exec("INSERT INTO tasks (id, title, status, is_archived) VALUES (2, 'Done Task', 'done', 0)")
	database.DB.Exec("INSERT INTO tasks (id, title, status, is_archived) VALUES (3, 'Archived Task', 'todo', 1)")
	database.DB.Exec("INSERT INTO tasks (id, title, status, is_archived) VALUES (4, 'Blocked Task', 'todo', 0)")
	database.DB.Exec("INSERT INTO task_obstacles (task_id, description, status) VALUES (4, 'Some obstacle', 'open')")

	// Testing the logic of the where clause construction
	getWhereClause := func(filter string) string {
		whereClause := "WHERE is_archived = 0"
		switch filter {
		case "done":
			whereClause = "WHERE status = 'done' AND is_archived = 0"
		case "archived":
			whereClause = "WHERE is_archived = 1"
		case "kendala":
			whereClause = "WHERE (SELECT COUNT(*) FROM task_obstacles WHERE task_id = t.id AND status = 'open') > 0 AND is_archived = 0"
		default:
			whereClause = "WHERE status != 'done' AND is_archived = 0"
		}
		return whereClause
	}

	tests := []struct {
		filter string
		want   int
	}{
		{"", 2},         // Active (Tasks 1 and 4)
		{"done", 1},     // Done (Task 2)
		{"archived", 1}, // Archived (Task 3)
		{"kendala", 1},  // Kendala (Task 4)
	}

	for _, tt := range tests {
		where := getWhereClause(tt.filter)
		query := "SELECT COUNT(*) FROM tasks t " + where
		var count int
		err := database.DB.QueryRow(query).Scan(&count)
		if err != nil {
			t.Errorf("filter %s: query failed: %v", tt.filter, err)
			continue
		}
		if count != tt.want {
			t.Errorf("filter %s: expected %d tasks, got %d", tt.filter, tt.want, count)
		}
	}
}
