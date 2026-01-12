package handlers

import (
	"ac_tracker/internal/database"
	"ac_tracker/internal/models"
	"html/template"
	"log"
	"net/http"
	"time"
)

func WorkDashboard(w http.ResponseWriter, r *http.Request) {
	// Filter & Search Params
	filter := r.URL.Query().Get("filter") // active (default), done, archived
	search := r.URL.Query().Get("q")

	whereClause := "WHERE is_archived = 0"
	if filter == "done" {
		whereClause = "WHERE status = 'done' AND is_archived = 0"
	} else if filter == "archived" {
		whereClause = "WHERE is_archived = 1"
	} else {
		// default active
		whereClause = "WHERE status != 'done' AND is_archived = 0"
	}

	args := []interface{}{}
	if search != "" {
		whereClause += " AND (title LIKE ? OR outcome LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	// Fetch with all timestamp fields and open obstacles count
	query := `
		SELECT t.id, t.title, COALESCE(t.outcome, '') as outcome, t.status, t.priority, t.is_archived, t.created_at, t.updated_at, t.completed_at,
		(SELECT COUNT(*) FROM task_obstacles WHERE task_id = t.id AND status = 'open') as open_obstacles
		FROM tasks t ` + whereClause + ` ORDER BY 
		CASE WHEN t.priority = 'high' THEN 1 WHEN t.priority = 'medium' THEN 2 ELSE 3 END, 
		t.created_at DESC`
	
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	loc, _ := time.LoadLocation("Asia/Jakarta")
	for rows.Next() {
		var t models.Task
		// Scan new columns priority and is_archived
		if err := rows.Scan(&t.ID, &t.Title, &t.Outcome, &t.Status, &t.Priority, &t.IsArchived, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt, &t.OpenObstaclesCount); err != nil {
			// Handle NULL priority/is_archived if legacy rows exist without default (though ALTER sets default)
			log.Println("Error scanning task:", err)
			continue
		}
		// Convert to local time
		t.CreatedAt = t.CreatedAt.In(loc)
		t.UpdatedAt = t.UpdatedAt.In(loc)
		if t.CompletedAt != nil {
			localComp := t.CompletedAt.In(loc)
			t.CompletedAt = &localComp
		}
		tasks = append(tasks, t)
	}

	tmpl, err := template.ParseFiles("templates/work_simple.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username := "Guest"
	role := ""
	if c, err := r.Cookie("user_name"); err == nil {
		username = c.Value
	}
	if c, err := r.Cookie("user_role"); err == nil {
		role = c.Value
	}

	tmpl.Execute(w, map[string]interface{}{
		"Tasks":    tasks,
		"Filter":   filter,
		"Search":   search,
		"Year":     time.Now().Year(),
		"UserName": username,
		"UserRole": role,
	})
}

func CreateSimpleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Task required", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec("INSERT INTO tasks(title, outcome, status, priority, is_archived, created_at, updated_at) VALUES(?, '', 'todo', 'medium', 0, datetime('now', 'localtime'), datetime('now', 'localtime'))", title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/work", http.StatusSeeOther)
}

func ToggleTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var currentStatus string
	err := database.DB.QueryRow("SELECT status FROM tasks WHERE id = ?", id).Scan(&currentStatus)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	newStatus := "done"
	var completedAt interface{} = "datetime('now', 'localtime')"
	if currentStatus == "done" {
		newStatus = "todo"
		completedAt = nil
	}

	if completedAt == nil {
		_, err = database.DB.Exec("UPDATE tasks SET status = ?, completed_at = NULL, updated_at = datetime('now', 'localtime') WHERE id = ?", newStatus, id)
	} else {
		_, err = database.DB.Exec("UPDATE tasks SET status = ?, completed_at = datetime('now', 'localtime'), updated_at = datetime('now', 'localtime') WHERE id = ?", newStatus, id)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/work", http.StatusSeeOther)
}

func UpdateTaskNotes(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	notes := r.FormValue("notes")

	_, err := database.DB.Exec("UPDATE tasks SET outcome = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", notes, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/work", http.StatusSeeOther)
}

func EditSimpleTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		notes := r.FormValue("notes")
		status := r.FormValue("status")
		priority := r.FormValue("priority")

		compAt := ""
		if status == "done" {
			compAt = ", completed_at = COALESCE(completed_at, datetime('now', 'localtime'))"
		} else {
			compAt = ", completed_at = NULL"
		}

		query := "UPDATE tasks SET title = ?, outcome = ?, status = ?, priority = ?, updated_at = datetime('now', 'localtime')" + compAt + " WHERE id = ?"
		_, err := database.DB.Exec(query, title, notes, status, priority, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/work", http.StatusSeeOther)
		return
	}

	var t models.Task
	err := database.DB.QueryRow("SELECT id, title, COALESCE(outcome, '') as outcome, status, priority FROM tasks WHERE id = ?", id).Scan(&t.ID, &t.Title, &t.Outcome, &t.Status, &t.Priority)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Fetch obstacles
	obsRows, err := database.DB.Query("SELECT id, description, status, created_at, resolved_at FROM task_obstacles WHERE task_id = ? ORDER BY created_at ASC", id)
	if err == nil {
		defer obsRows.Close()
		loc, _ := time.LoadLocation("Asia/Jakarta")
		for obsRows.Next() {
			var o models.Obstacle
			obsRows.Scan(&o.ID, &o.Description, &o.Status, &o.CreatedAt, &o.ResolvedAt)
			o.CreatedAt = o.CreatedAt.In(loc)
			if o.ResolvedAt != nil {
				localRes := o.ResolvedAt.In(loc)
				o.ResolvedAt = &localRes
			}
			t.Obstacles = append(t.Obstacles, o)
		}
	}

	tmpl, err := template.ParseFiles("templates/work_edit.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username := "Guest"
	role := ""
	if c, err := r.Cookie("user_name"); err == nil {
		username = c.Value
	}
	if c, err := r.Cookie("user_role"); err == nil {
		role = c.Value
	}

	tmpl.Execute(w, map[string]interface{}{
		"Task":     t,
		"Year":     time.Now().Year(),
		"UserName": username,
		"UserRole": role,
	})
}

func AddObstacle(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	description := r.FormValue("description")
	if description == "" {
		http.Error(w, "Description required", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec("INSERT INTO task_obstacles(task_id, description, status, created_at) VALUES(?, ?, 'open', datetime('now', 'localtime'))", taskID, description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/work/edit/"+taskID, http.StatusSeeOther)
}

func ResolveObstacle(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	obsID := r.PathValue("obsId")

	var currentStatus string
	database.DB.QueryRow("SELECT status FROM task_obstacles WHERE id = ?", obsID).Scan(&currentStatus)

	newStatus := "resolved"
	var resolvedAt interface{} = "datetime('now', 'localtime')"
	if currentStatus == "resolved" {
		newStatus = "open"
		resolvedAt = nil
	}

	if resolvedAt == nil {
		database.DB.Exec("UPDATE task_obstacles SET status = ?, resolved_at = NULL WHERE id = ?", newStatus, obsID)
	} else {
		database.DB.Exec("UPDATE task_obstacles SET status = ?, resolved_at = datetime('now', 'localtime') WHERE id = ?", newStatus, obsID)
	}

	http.Redirect(w, r, "/work/edit/"+taskID, http.StatusSeeOther)
}

func DeleteObstacle(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	obsID := r.PathValue("obsId")

	database.DB.Exec("DELETE FROM task_obstacles WHERE id = ?", obsID)
	http.Redirect(w, r, "/work/edit/"+taskID, http.StatusSeeOther)
}

func ArchiveTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	database.DB.Exec("UPDATE tasks SET is_archived = 1 WHERE id = ?", id)
	http.Redirect(w, r, "/work", http.StatusSeeOther)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// Hard delete task and its obstacles
	database.DB.Exec("DELETE FROM task_obstacles WHERE task_id = ?", id)
	database.DB.Exec("DELETE FROM tasks WHERE id = ?", id)
	http.Redirect(w, r, "/work", http.StatusSeeOther)
}
