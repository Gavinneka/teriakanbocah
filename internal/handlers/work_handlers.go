package handlers

import (
	"ac_tracker/internal/database"
	"ac_tracker/internal/models"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

func WorkDashboard(w http.ResponseWriter, r *http.Request) {
	// Filter & Search Params
	filter := r.URL.Query().Get("filter") // active (default), done, archived
	search := r.URL.Query().Get("q")

	whereClause := "WHERE is_archived = 0"
	switch filter {
	case "done":
		whereClause = "WHERE status = 'done' AND is_archived = 0"
	case "kendala":
		whereClause = "WHERE (SELECT COUNT(*) FROM task_details WHERE task_id = t.id AND COALESCE(obstacle, '') != '' AND COALESCE(obstacle_resolved, 0) = 0) > 0 AND is_archived = 0"
	default:
		// default active
		whereClause = "WHERE status != 'done' AND is_archived = 0"
	}

	args := []interface{}{}
	if search != "" {
		whereClause += " AND (title LIKE ? OR outcome LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	// Fetch with all timestamp fields and open obstacles count from task_details
	query := `
		SELECT t.id, t.title, COALESCE(t.outcome, '') as outcome, t.status, t.priority, t.is_archived, t.created_at, t.updated_at, t.completed_at, COALESCE(t.assigned_to, '') as assigned_to, t.due_date,
		(SELECT COUNT(*) FROM task_details WHERE task_id = t.id AND COALESCE(obstacle, '') != '' AND COALESCE(obstacle_resolved, 0) = 0) as open_obstacles
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

	// Map to store obstacles by TaskID for the list
	obstacleMap := make(map[int][]models.Obstacle)

	// Map to store task details by TaskID for the list
	detailMap := make(map[int][]models.TaskDetail)

	// Slice for the global summary at the top
	type ObstacleSummary struct {
		TaskTitle   string
		Description string
	}
	var globalObstacles []ObstacleSummary

	// Fetch ALL task details for dashboard row display
	detailRows, err := database.DB.Query(`
		SELECT id, task_id, description, COALESCE(progress, '') as progress, COALESCE(obstacle, '') as obstacle, is_done, COALESCE(obstacle_resolved, 0) as obstacle_resolved, created_at, updated_at
		FROM task_details
		ORDER BY created_at ASC`)
	if err == nil {
		defer detailRows.Close()
		for detailRows.Next() {
			var d models.TaskDetail
			var tID int
			if err := detailRows.Scan(&d.ID, &tID, &d.Description, &d.Progress, &d.Obstacle, &d.IsDone, &d.ObstacleResolved, &d.CreatedAt, &d.UpdatedAt); err == nil {
				d.CreatedAt = d.CreatedAt.In(loc)
				d.UpdatedAt = d.UpdatedAt.In(loc)
				detailMap[tID] = append(detailMap[tID], d)
			}
		}
	}

	// Fetch ALL open obstacles from task_details
	obsRows, err := database.DB.Query(`
		SELECT d.id, d.task_id, d.description, d.obstacle, t.title
		FROM task_details d
		JOIN tasks t ON d.task_id = t.id
		WHERE COALESCE(d.obstacle, '') != '' AND COALESCE(d.obstacle_resolved, 0) = 0
		ORDER BY d.created_at DESC`)

	if err == nil {
		defer obsRows.Close()
		for obsRows.Next() {
			var o models.Obstacle
			var tID int
			var detailDesc string
			var tTitle string

			if err := obsRows.Scan(&o.ID, &tID, &detailDesc, &o.Description, &tTitle); err == nil {
				// Prepend detail item description so it's clear
				o.Description = "[" + detailDesc + "] " + o.Description
				o.Status = "open"

				// Add to map for per-row display
				obstacleMap[tID] = append(obstacleMap[tID], o)

				// Add to slice for top summary
				globalObstacles = append(globalObstacles, ObstacleSummary{
					TaskTitle:   tTitle,
					Description: o.Description,
				})
			}
		}
	}

	for rows.Next() {
		var t models.Task
		var dueTime sql.NullTime
		// Scan new columns priority, is_archived, assigned_to, due_date
		if err := rows.Scan(&t.ID, &t.Title, &t.Outcome, &t.Status, &t.Priority, &t.IsArchived, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt, &t.AssignedTo, &dueTime, &t.OpenObstaclesCount); err != nil {
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
		if dueTime.Valid {
			localDue := dueTime.Time.In(loc)
			t.DueDate = &localDue
		}

		// Assign obstacles from map (In-memory join)
		if obs, ok := obstacleMap[t.ID]; ok {
			t.Obstacles = obs
		}

		// Assign details from map (In-memory join)
		if details, ok := detailMap[t.ID]; ok {
			t.Details = details
		}

		tasks = append(tasks, t)
	}

	tmpl, err := template.ParseFiles("templates/work_simple.html", "templates/sidebar_layout.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	base := GetBaseData(r)

	tmpl.Execute(w, map[string]interface{}{
		"Tasks":           tasks,
		"GlobalObstacles": globalObstacles,
		"Filter":          filter,
		"Search":          search,
		"Year":            base.Year,
		"UserName":        base.UserName,
		"UserRole":        base.UserRole,
	})
}

func CreateSimpleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	dueDate := r.FormValue("due_date")

	if title == "" {
		http.Error(w, "Task required", http.StatusBadRequest)
		return
	}

	var due interface{}
	if dueDate != "" {
		due = dueDate
	}

	_, err := database.DB.Exec("INSERT INTO tasks(title, outcome, status, priority, assigned_to, due_date, is_archived, created_at, updated_at) VALUES(?, '', 'todo', 'medium', '', ?, 0, datetime('now', 'localtime'), datetime('now', 'localtime'))", title, due)
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
		dueDate := r.FormValue("due_date")

		var due interface{}
		if dueDate != "" {
			due = dueDate
		}

		compAt := ""
		if status == "done" {
			compAt = ", completed_at = COALESCE(completed_at, datetime('now', 'localtime'))"
		} else {
			compAt = ", completed_at = NULL"
		}

		query := "UPDATE tasks SET title = ?, outcome = ?, status = ?, priority = ?, due_date = ?, updated_at = datetime('now', 'localtime')" + compAt + " WHERE id = ?"
		_, err := database.DB.Exec(query, title, notes, status, priority, due, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/work", http.StatusSeeOther)
		return
	}

	var t models.Task
	var dueTime sql.NullTime
	err := database.DB.QueryRow("SELECT id, title, COALESCE(outcome, '') as outcome, status, priority, COALESCE(assigned_to, '') as assigned_to, due_date FROM tasks WHERE id = ?", id).Scan(&t.ID, &t.Title, &t.Outcome, &t.Status, &t.Priority, &t.AssignedTo, &dueTime)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if dueTime.Valid {
		loc, _ := time.LoadLocation("Asia/Jakarta")
		localDue := dueTime.Time.In(loc)
		t.DueDate = &localDue
	}

	// Fetch task details
	detailRows, err := database.DB.Query("SELECT id, description, progress, obstacle, is_done, COALESCE(obstacle_resolved, 0), created_at, updated_at FROM task_details WHERE task_id = ? ORDER BY created_at ASC", id)
	if err == nil {
		defer detailRows.Close()
		loc, _ := time.LoadLocation("Asia/Jakarta")
		for detailRows.Next() {
			var d models.TaskDetail
			detailRows.Scan(&d.ID, &d.Description, &d.Progress, &d.Obstacle, &d.IsDone, &d.ObstacleResolved, &d.CreatedAt, &d.UpdatedAt)
			d.CreatedAt = d.CreatedAt.In(loc)
			d.UpdatedAt = d.UpdatedAt.In(loc)
			t.Details = append(t.Details, d)
		}
	}

	tmpl, err := template.ParseFiles("templates/work_edit.html", "templates/sidebar_layout.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	base := GetBaseData(r)
	if base.UserName == "Guest" {
		// Optional: Redirect if strict, or just render "Guest"
	}

	tmpl.Execute(w, map[string]interface{}{
		"Task":     t,
		"Year":     base.Year,
		"UserName": base.UserName,
		"UserRole": base.UserRole,
	})
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// Hard delete task, its obstacles, and details
	database.DB.Exec("DELETE FROM task_obstacles WHERE task_id = ?", id)
	database.DB.Exec("DELETE FROM task_details WHERE task_id = ?", id)
	database.DB.Exec("DELETE FROM tasks WHERE id = ?", id)
	http.Redirect(w, r, "/work", http.StatusSeeOther)
}

func AddTaskDetail(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	description := r.FormValue("description")
	progress := r.FormValue("progress")
	obstacle := r.FormValue("obstacle")
	redirect := r.FormValue("redirect")

	if description == "" {
		http.Error(w, "Description required", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec("INSERT INTO task_details(task_id, description, progress, obstacle, created_at, updated_at) VALUES(?, ?, ?, ?, datetime('now', 'localtime'), datetime('now', 'localtime'))", taskID, description, progress, obstacle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if redirect == "drawer" {
		http.Redirect(w, r, "/work/tasks/"+taskID+"/drawer", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/work/edit/"+taskID, http.StatusSeeOther)
	}
}

func EditTaskDetail(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	detailID := r.PathValue("detailId")
	description := r.FormValue("description")
	progress := r.FormValue("progress")
	obstacle := r.FormValue("obstacle")
	redirect := r.FormValue("redirect")

	if description == "" {
		http.Error(w, "Description required", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec("UPDATE task_details SET description = ?, progress = ?, obstacle = ?, updated_at = datetime('now', 'localtime') WHERE id = ? AND task_id = ?", description, progress, obstacle, detailID, taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if redirect == "drawer" {
		http.Redirect(w, r, "/work/tasks/"+taskID+"/drawer", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/work/edit/"+taskID, http.StatusSeeOther)
	}
}

func DeleteTaskDetail(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	detailID := r.PathValue("detailId")
	redirect := r.FormValue("redirect")

	_, err := database.DB.Exec("DELETE FROM task_details WHERE id = ? AND task_id = ?", detailID, taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if redirect == "drawer" {
		http.Redirect(w, r, "/work/tasks/"+taskID+"/drawer", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/work/edit/"+taskID, http.StatusSeeOther)
	}
}

func ToggleTaskDetail(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	detailID := r.PathValue("detailId")
	redirectTo := r.FormValue("redirect")

	var isDone int
	err := database.DB.QueryRow("SELECT is_done FROM task_details WHERE id = ? AND task_id = ?", detailID, taskID).Scan(&isDone)
	if err != nil {
		http.Error(w, "Detail not found", http.StatusNotFound)
		return
	}

	newDone := 0
	if isDone == 0 {
		newDone = 1
	}

	_, err = database.DB.Exec("UPDATE task_details SET is_done = ?, updated_at = datetime('now', 'localtime') WHERE id = ? AND task_id = ?", newDone, detailID, taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if redirectTo == "drawer" {
		http.Redirect(w, r, "/work/tasks/"+taskID+"/drawer", http.StatusSeeOther)
	} else if redirectTo == "edit" {
		http.Redirect(w, r, "/work/edit/"+taskID, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/work", http.StatusSeeOther)
	}
}

func ToggleObstacle(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	detailID := r.PathValue("detailId")
	redirectTo := r.FormValue("redirect")

	var current int
	err := database.DB.QueryRow("SELECT COALESCE(obstacle_resolved, 0) FROM task_details WHERE id = ? AND task_id = ?", detailID, taskID).Scan(&current)
	if err != nil {
		http.Error(w, "Detail not found", http.StatusNotFound)
		return
	}

	newVal := 0
	if current == 0 {
		newVal = 1
	}

	_, err = database.DB.Exec("UPDATE task_details SET obstacle_resolved = ?, updated_at = datetime('now', 'localtime') WHERE id = ? AND task_id = ?", newVal, detailID, taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if redirectTo == "drawer" {
		http.Redirect(w, r, "/work/tasks/"+taskID+"/drawer", http.StatusSeeOther)
	} else if redirectTo == "edit" {
		http.Redirect(w, r, "/work/edit/"+taskID, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/work", http.StatusSeeOther)
	}
}

func WorkInbox(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(`
		SELECT id, title, COALESCE(outcome, '') as outcome, status, priority, is_archived, created_at, updated_at, completed_at, COALESCE(assigned_to, '') as assigned_to, due_date
		FROM tasks
		WHERE status = 'inbox' AND is_archived = 0
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	loc, _ := time.LoadLocation("Asia/Jakarta")

	for rows.Next() {
		var t models.Task
		var dueTime sql.NullTime
		if err := rows.Scan(&t.ID, &t.Title, &t.Outcome, &t.Status, &t.Priority, &t.IsArchived, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt, &t.AssignedTo, &dueTime); err != nil {
			log.Println("Error scanning task in inbox:", err)
			continue
		}
		t.CreatedAt = t.CreatedAt.In(loc)
		t.UpdatedAt = t.UpdatedAt.In(loc)
		if t.CompletedAt != nil {
			localComp := t.CompletedAt.In(loc)
			t.CompletedAt = &localComp
		}
		if dueTime.Valid {
			localDue := dueTime.Time.In(loc)
			t.DueDate = &localDue
		}
		tasks = append(tasks, t)
	}

	tmpl, err := template.ParseFiles("templates/work_inbox.html", "templates/sidebar_layout.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	base := GetBaseData(r)
	tmpl.Execute(w, map[string]interface{}{
		"Tasks":    tasks,
		"Year":     base.Year,
		"UserName": base.UserName,
		"UserRole": base.UserRole,
	})
}

func CreateInboxTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`
		INSERT INTO tasks(title, outcome, status, priority, assigned_to, is_archived, created_at, updated_at)
		VALUES(?, '', 'inbox', 'medium', '', 0, datetime('now', 'localtime'), datetime('now', 'localtime'))
	`, title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/work/inbox", http.StatusSeeOther)
}

func TaskDrawer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var t models.Task
	var dueTime sql.NullTime
	err := database.DB.QueryRow(`
		SELECT id, title, COALESCE(outcome, '') as outcome, COALESCE(estimate, 0) as estimate, status, priority, due_date, created_at, updated_at
		FROM tasks
		WHERE id = ?
	`, id).Scan(&t.ID, &t.Title, &t.Outcome, &t.Estimate, &t.Status, &t.Priority, &dueTime, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if dueTime.Valid {
		loc, _ := time.LoadLocation("Asia/Jakarta")
		localDue := dueTime.Time.In(loc)
		t.DueDate = &localDue
	}

	// Fetch task details
	detailRows, err := database.DB.Query("SELECT id, description, progress, obstacle, is_done, COALESCE(obstacle_resolved, 0), created_at, updated_at FROM task_details WHERE task_id = ? ORDER BY created_at ASC", id)
	if err == nil {
		defer detailRows.Close()
		loc, _ := time.LoadLocation("Asia/Jakarta")
		for detailRows.Next() {
			var d models.TaskDetail
			detailRows.Scan(&d.ID, &d.Description, &d.Progress, &d.Obstacle, &d.IsDone, &d.ObstacleResolved, &d.CreatedAt, &d.UpdatedAt)
			d.CreatedAt = d.CreatedAt.In(loc)
			d.UpdatedAt = d.UpdatedAt.In(loc)
			t.Details = append(t.Details, d)
		}
	}

	tmpl, err := template.ParseFiles("templates/task_drawer_partial.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, map[string]interface{}{
		"Task": t,
	})
}

func UpdateTaskDrawer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	title := r.FormValue("title")
	outcome := r.FormValue("outcome")
	status := r.FormValue("status")
	priority := r.FormValue("priority")
	dueDateStr := r.FormValue("due_date")
	estimateStr := r.FormValue("estimate")

	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	estimate, _ := strconv.Atoi(estimateStr)

	var dueDate interface{}
	if dueDateStr != "" {
		dVal, err := time.Parse("2006-01-02", dueDateStr)
		if err == nil {
			dueDate = dVal
		} else {
			dueDate = nil
		}
	} else {
		dueDate = nil
	}

	compAt := ""
	if status == "done" {
		compAt = ", completed_at = COALESCE(completed_at, datetime('now', 'localtime'))"
	} else {
		compAt = ", completed_at = NULL"
	}

	query := "UPDATE tasks SET title = ?, outcome = ?, status = ?, priority = ?, due_date = ?, estimate = ?, updated_at = datetime('now', 'localtime')" + compAt + " WHERE id = ?"
	_, err := database.DB.Exec(query, title, outcome, status, priority, dueDate, estimate, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "taskUpdated")
	w.WriteHeader(http.StatusOK)
}
