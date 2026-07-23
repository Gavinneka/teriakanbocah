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

func ImprovementList(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter") // all, selesai, proses, ditunda

	whereClause := ""
	switch filter {
	case "selesai":
		whereClause = "WHERE status = 'selesai'"
	case "proses":
		whereClause = "WHERE status = 'proses'"
	case "ditunda":
		whereClause = "WHERE status = 'ditunda'"
	default:
		filter = "all"
		whereClause = ""
	}

	query := `SELECT id, title, cost, status, COALESCE(notes, '') as notes, done_at, created_at, updated_at
		FROM improvement_logs ` + whereClause + ` ORDER BY created_at DESC`

	rows, err := database.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []models.ImprovementLog
	loc, _ := time.LoadLocation("Asia/Jakarta")

	for rows.Next() {
		var item models.ImprovementLog
		var doneAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.Title, &item.Cost, &item.Status, &item.Notes, &doneAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			log.Println("Error scanning improvement_log:", err)
			continue
		}
		item.CreatedAt = item.CreatedAt.In(loc)
		item.UpdatedAt = item.UpdatedAt.In(loc)
		if doneAt.Valid {
			localDone := doneAt.Time.In(loc)
			item.DoneAt = &localDone
		}
		items = append(items, item)
	}

	// Hitung summary total biaya
	var totalCost int
	for _, item := range items {
		totalCost += item.Cost
	}

	// Total keseluruhan (tanpa filter) untuk summary card
	var totalAll int
	var countAll int
	database.DB.QueryRow("SELECT COUNT(*), COALESCE(SUM(cost), 0) FROM improvement_logs").Scan(&countAll, &totalAll)

	tmpl, err := template.New("improvement.html").Funcs(template.FuncMap{
		"formatRupiah": func(n int) string {
			s := strconv.Itoa(n)
			result := ""
			for i, c := range s {
				if i > 0 && (len(s)-i)%3 == 0 {
					result += "."
				}
				result += string(c)
			}
			return "Rp " + result
		},
	}).ParseFiles("templates/improvement.html", "templates/sidebar_layout.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	base := GetBaseData(r)
	if err := tmpl.Execute(w, map[string]interface{}{
		"Items":     items,
		"Filter":    filter,
		"TotalCost": totalCost,
		"TotalAll":  totalAll,
		"CountAll":  countAll,
		"Year":      base.Year,
		"UserName":  base.UserName,
		"UserRole":  base.UserRole,
	}); err != nil {
		log.Println("Error rendering improvement list:", err)
	}
}

func ImprovementAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	costStr := r.FormValue("cost")
	status := r.FormValue("status")
	notes := r.FormValue("notes")
	doneAtStr := r.FormValue("done_at")

	if title == "" {
		http.Error(w, "Nama improvement wajib diisi", http.StatusBadRequest)
		return
	}

	cost, _ := strconv.Atoi(costStr)
	if status == "" {
		status = "selesai"
	}

	var doneAt interface{}
	if doneAtStr != "" {
		doneAt = doneAtStr
	}

	_, err := database.DB.Exec(
		`INSERT INTO improvement_logs(title, cost, status, notes, done_at, created_at, updated_at)
		 VALUES(?, ?, ?, ?, ?, datetime('now', 'localtime'), datetime('now', 'localtime'))`,
		title, cost, status, notes, doneAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/improvement", http.StatusSeeOther)
}

func ImprovementEdit(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		costStr := r.FormValue("cost")
		status := r.FormValue("status")
		notes := r.FormValue("notes")
		doneAtStr := r.FormValue("done_at")

		cost, _ := strconv.Atoi(costStr)

		var doneAt interface{}
		if doneAtStr != "" {
			doneAt = doneAtStr
		}

		_, err := database.DB.Exec(
			`UPDATE improvement_logs SET title = ?, cost = ?, status = ?, notes = ?, done_at = ?, updated_at = datetime('now', 'localtime') WHERE id = ?`,
			title, cost, status, notes, doneAt, id,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/improvement", http.StatusSeeOther)
		return
	}

	// GET: tampilkan form edit
	var item models.ImprovementLog
	var doneAt sql.NullTime
	err := database.DB.QueryRow(
		`SELECT id, title, cost, status, COALESCE(notes, '') as notes, done_at, created_at, updated_at FROM improvement_logs WHERE id = ?`, id,
	).Scan(&item.ID, &item.Title, &item.Cost, &item.Status, &item.Notes, &doneAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		http.Error(w, "Item tidak ditemukan", http.StatusNotFound)
		return
	}
	if doneAt.Valid {
		loc, _ := time.LoadLocation("Asia/Jakarta")
		localDone := doneAt.Time.In(loc)
		item.DoneAt = &localDone
	}

	tmpl, err := template.New("improvement_edit.html").Funcs(template.FuncMap{
		"formatRupiah": func(n int) string {
			s := strconv.Itoa(n)
			result := ""
			for i, c := range s {
				if i > 0 && (len(s)-i)%3 == 0 {
					result += "."
				}
				result += string(c)
			}
			return "Rp " + result
		},
	}).ParseFiles("templates/improvement_edit.html", "templates/sidebar_layout.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	base := GetBaseData(r)
	if err := tmpl.Execute(w, map[string]interface{}{
		"Item":     item,
		"Year":     base.Year,
		"UserName": base.UserName,
		"UserRole": base.UserRole,
	}); err != nil {
		log.Println("Error rendering improvement edit:", err)
	}
}

func ImprovementDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	database.DB.Exec("DELETE FROM improvement_logs WHERE id = ?", id)
	http.Redirect(w, r, "/improvement", http.StatusSeeOther)
}
