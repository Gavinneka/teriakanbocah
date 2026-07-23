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
		is_archived BOOLEAN DEFAULT 0,
		assigned_to TEXT DEFAULT '',
		due_date DATETIME
	);
	CREATE TABLE task_obstacles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER,
		description TEXT,
		status TEXT DEFAULT 'open',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		resolved_at DATETIME
	);
	CREATE TABLE task_details (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER,
		description TEXT NOT NULL,
		progress TEXT DEFAULT '',
		obstacle TEXT DEFAULT '',
		obstacle_resolved INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
func TestWorkDashboardFilteringQueries(t *testing.T) {
	setupTestDB()
	defer database.DB.Close()

	// Seed data
	database.DB.Exec("INSERT INTO tasks (id, title, status, is_archived) VALUES (1, 'Active Task', 'todo', 0)")
	database.DB.Exec("INSERT INTO tasks (id, title, status, is_archived) VALUES (2, 'Done Task', 'done', 0)")
	database.DB.Exec("INSERT INTO tasks (id, title, status, is_archived) VALUES (3, 'Archived Task', 'todo', 1)")
	database.DB.Exec("INSERT INTO tasks (id, title, status, is_archived) VALUES (4, 'Blocked Task', 'todo', 0)")
	database.DB.Exec("INSERT INTO task_details (task_id, description, obstacle) VALUES (4, 'Some detail', 'Some obstacle')")
	database.DB.Exec("INSERT INTO tasks (id, title, status, is_archived) VALUES (5, 'Resolved Task', 'todo', 0)")
	database.DB.Exec("INSERT INTO task_details (task_id, description, obstacle, obstacle_resolved) VALUES (5, 'Resolved detail', 'Resolved obstacle', 1)")

	// Testing the logic of the where clause construction
	getWhereClause := func(filter string) string {
		whereClause := "WHERE is_archived = 0"
		switch filter {
		case "done":
			whereClause = "WHERE status = 'done' AND is_archived = 0"
		case "kendala":
			whereClause = "WHERE (SELECT COUNT(*) FROM task_details WHERE task_id = t.id AND COALESCE(obstacle, '') != '' AND COALESCE(obstacle_resolved, 0) = 0) > 0 AND is_archived = 0"
		default:
			whereClause = "WHERE status != 'done' AND is_archived = 0"
		}
		return whereClause
	}

	tests := []struct {
		filter string
		want   int
	}{
		{"", 3},        // Active (Tasks 1, 4, and 5)
		{"done", 1},    // Done (Task 2)
		{"kendala", 1}, // Only Task 4; Task 5's obstacle is resolved
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

func TestAddTaskDetail(t *testing.T) {
	setupTestDB()
	defer database.DB.Close()

	database.DB.Exec("INSERT INTO tasks (id, title) VALUES (1, 'Task 1')")

	form := url.Values{}
	form.Add("description", "Detail 1")
	form.Add("progress", "In progress")
	form.Add("obstacle", "Some obstacle")
	req := httptest.NewRequest("POST", "/work/tasks/1/details", strings.NewReader(form.Encode()))
	req.SetPathValue("id", "1")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	AddTaskDetail(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", rr.Code)
	}

	var desc, prog, obs string
	err := database.DB.QueryRow("SELECT description, progress, obstacle FROM task_details WHERE task_id = 1").Scan(&desc, &prog, &obs)
	if err != nil {
		t.Fatal(err)
	}
	if desc != "Detail 1" || prog != "In progress" || obs != "Some obstacle" {
		t.Errorf("unexpected detail values: %s, %s, %s", desc, prog, obs)
	}
}

func TestEditTaskDetail(t *testing.T) {
	setupTestDB()
	defer database.DB.Close()

	database.DB.Exec("INSERT INTO tasks (id, title) VALUES (1, 'Task 1')")
	database.DB.Exec("INSERT INTO task_details (id, task_id, description, progress, obstacle) VALUES (10, 1, 'Old Desc', 'Old Prog', 'Old Obs')")

	form := url.Values{}
	form.Add("description", "New Desc")
	form.Add("progress", "New Prog")
	form.Add("obstacle", "New Obs")
	req := httptest.NewRequest("POST", "/work/tasks/1/details/10/edit", strings.NewReader(form.Encode()))
	req.SetPathValue("id", "1")
	req.SetPathValue("detailId", "10")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	EditTaskDetail(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", rr.Code)
	}

	var desc, prog, obs string
	err := database.DB.QueryRow("SELECT description, progress, obstacle FROM task_details WHERE id = 10").Scan(&desc, &prog, &obs)
	if err != nil {
		t.Fatal(err)
	}
	if desc != "New Desc" || prog != "New Prog" || obs != "New Obs" {
		t.Errorf("unexpected detail values after edit: %s, %s, %s", desc, prog, obs)
	}
}

func TestDeleteTaskDetail(t *testing.T) {
	setupTestDB()
	defer database.DB.Close()

	database.DB.Exec("INSERT INTO tasks (id, title) VALUES (1, 'Task 1')")
	database.DB.Exec("INSERT INTO task_details (id, task_id, description) VALUES (10, 1, 'To delete')")

	req := httptest.NewRequest("POST", "/work/tasks/1/details/10/delete", nil)
	req.SetPathValue("id", "1")
	req.SetPathValue("detailId", "10")
	rr := httptest.NewRecorder()

	DeleteTaskDetail(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", rr.Code)
	}

	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM task_details WHERE id = 10").Scan(&count)
	if count != 0 {
		t.Errorf("expected detail to be deleted, but it still exists")
	}
}
