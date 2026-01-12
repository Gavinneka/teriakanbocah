package handlers

import (
	"ac_tracker/internal/database"
	"ac_tracker/internal/models"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func GetRecords(w http.ResponseWriter, r *http.Request) {
	if !CheckPermission(r, "ac", "view") {
		http.Error(w, "Forbidden: No View Permission", http.StatusForbidden)
		return
	}
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	room := r.URL.Query().Get("room")

	query := "SELECT id, room, unit, activity, date, status, notes FROM maintenance_records WHERE 1=1"
	args := []interface{}{}

	if startDate != "" {
		query += " AND date >= ?"
		// Parse and ensure start of day? SQLite datetime is string or numeric.
		// Assuming we stored as Time, driver usually saves as string "2006-01-02 15:04:05...".
		// We just compare string prefixes or use date function if needed.
		// "2025-01-01" matches >= "2025-01-01 00:00:00" if standard format.
		// Let's safely cast to time object for the driver if possible, or string.
		// Go's sqlite driver handles time.Time -> string ISO8601.
		// Input form returns "YYYY-MM-DD".
		// Let's convert to time.Time to be safe with the driver.
		t, _ := time.Parse("2006-01-02", startDate)
		args = append(args, t)
	}

	if endDate != "" {
		query += " AND date <= ?"
		// Add 23:59:59 to include the end date fully?
		t, _ := time.Parse("2006-01-02", endDate)
		t = t.Add(24 * time.Hour).Add(-1 * time.Second)
		args = append(args, t)
	}

	if room != "" {
		query += " AND room = ?"
		args = append(args, room)
	}

	query += " ORDER BY date desc"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []models.MaintenanceRecord
	for rows.Next() {
		var rec models.MaintenanceRecord
		if err := rows.Scan(&rec.ID, &rec.Room, &rec.Unit, &rec.Activity, &rec.Date, &rec.Status, &rec.Notes); err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		records = append(records, rec)
	}

	// Filter rooms for dropdown
	roomRows, err := database.DB.Query("SELECT DISTINCT room FROM maintenance_records ORDER BY room")
	var rooms []string
	if err == nil {
		defer roomRows.Close()
		for roomRows.Next() {
			var r string
			roomRows.Scan(&r)
			rooms = append(rooms, r)
		}
	}

	tmpl, err := template.ParseFiles(
		"templates/index.html",
		"templates/base.html",
		"templates/form.html",
		"templates/record_item.html",
	)
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

	if err := tmpl.Execute(w, map[string]interface{}{
		"Records":     records,
		"Rooms":       rooms,
		"StartDate":   startDate,
		"EndDate":     endDate,
		"FilterRoom":  room,
		"Year":        time.Now().Year(),
		"UserName":    username,
		"UserRole":    role,
	}); err != nil {
		log.Println("Error executing template:", err)
	}
}

func CreateRecord(w http.ResponseWriter, r *http.Request) {
	if !CheckPermission(r, "ac", "create") {
		http.Error(w, "Forbidden: No Create Permission", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dateStr := r.FormValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	room := r.FormValue("room_select")
	if room == "Lainnya" {
		room = r.FormValue("room_custom")
	}

	record := models.MaintenanceRecord{
		Room:     room,
		Unit:     r.FormValue("unit"),
		Activity: r.FormValue("activity"),
		Date:     date,
		Status:   r.FormValue("status"),
		Notes:    r.FormValue("notes"),
	}

	stmt, err := database.DB.Prepare("INSERT INTO maintenance_records(room, unit, activity, date, status, notes) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := stmt.Exec(record.Room, record.Unit, record.Activity, record.Date, record.Status, record.Notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	record.ID = int(id)

	// For HTMX: Render just the new row
	tmpl, err := template.ParseFiles("templates/record_item.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "record_item", record); err != nil {
		log.Println("Error executing template:", err)
	}
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("dashboard.html").Funcs(template.FuncMap{
		"contains": strings.Contains,
	}).ParseFiles("templates/dashboard.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username := ""
	role := ""
	allowedModules := ""
	if c, err := r.Cookie("user_name"); err == nil {
		username = c.Value
	}
	if c, err := r.Cookie("user_role"); err == nil {
		role = c.Value
	}
	if c, err := r.Cookie("user_modules"); err == nil {
		allowedModules = c.Value
	}

	tmpl.Execute(w, map[string]interface{}{
		"Year":           time.Now().Year(),
		"UserName":       username,
		"UserRole":       role,
		"AllowedModules": allowedModules,
	})
}

func DeleteRecord(w http.ResponseWriter, r *http.Request) {
	if !CheckPermission(r, "ac", "delete") {
		http.Error(w, "Forbidden: No Delete Permission", http.StatusForbidden)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	stmt, err := database.DB.Prepare("DELETE FROM maintenance_records WHERE id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // HTMX expects 200 OK to remove the element if we return empty body or configured swap
}

func EditRecord(w http.ResponseWriter, r *http.Request) {
	if !CheckPermission(r, "ac", "edit") {
		http.Error(w, "Forbidden: No Edit Permission", http.StatusForbidden)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	var rec models.MaintenanceRecord
	err := database.DB.QueryRow("SELECT id, room, unit, activity, date, status, notes FROM maintenance_records WHERE id = ?", id).Scan(
		&rec.ID, &rec.Room, &rec.Unit, &rec.Activity, &rec.Date, &rec.Status, &rec.Notes,
	)
	if err != nil {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	// Filter rooms for dropdown
	roomRows, err := database.DB.Query("SELECT DISTINCT room FROM maintenance_records ORDER BY room")
	var rooms []string
	if err == nil {
		defer roomRows.Close()
		for roomRows.Next() {
			var r string
			roomRows.Scan(&r)
			rooms = append(rooms, r)
		}
	}

	tmpl, err := template.ParseFiles("templates/edit_form.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Record": rec,
		"Rooms":  rooms,
	})
}

func UpdateRecord(w http.ResponseWriter, r *http.Request) {
	if !CheckPermission(r, "ac", "edit") {
		http.Error(w, "Forbidden: No Edit Permission", http.StatusForbidden)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	dateStr := r.FormValue("date")
	date, _ := time.Parse("2006-01-02", dateStr)

	room := r.FormValue("room_select")
	if room == "Lainnya" {
		room = r.FormValue("room_custom")
	}

	_, err := database.DB.Exec("UPDATE maintenance_records SET room=?, unit=?, activity=?, date=?, status=?, notes=? WHERE id=?",
		room, r.FormValue("unit"), r.FormValue("activity"), date, r.FormValue("status"), r.FormValue("notes"), id)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch updated record to render row
	var rec models.MaintenanceRecord
	database.DB.QueryRow("SELECT id, room, unit, activity, date, status, notes FROM maintenance_records WHERE id = ?", id).Scan(
		&rec.ID, &rec.Room, &rec.Unit, &rec.Activity, &rec.Date, &rec.Status, &rec.Notes,
	)

	tmpl, err := template.ParseFiles("templates/record_item.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, rec)
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/login.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Year": time.Now().Year(),
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	var dbPassword string
	var role string
	var userID int
	var isActive bool
	err := database.DB.QueryRow("SELECT id, password, role, COALESCE(is_active, 1) FROM users WHERE username = ?", username).Scan(&userID, &dbPassword, &role, &isActive)

	if err == nil && isActive {
		if bcrypt.ErrMismatchedHashAndPassword != bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password)) {
			// Login Success - Update last_login and log activity
			database.DB.Exec("UPDATE users SET last_login = datetime('now', 'localtime') WHERE id = ?", userID)
			
			// Log login activity
			ipAddress := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ipAddress = forwarded
			}
			database.DB.Exec("INSERT INTO login_logs(user_id, username, login_at, ip_address) VALUES(?, ?, datetime('now', 'localtime'), ?)", userID, username, ipAddress)
			
			http.SetCookie(w, &http.Cookie{
				Name:  "session_token",
				Value: "authenticated",
				Path:  "/",
			})
			http.SetCookie(w, &http.Cookie{
				Name:  "user_role",
				Value: role,
				Path:  "/",
			})
			http.SetCookie(w, &http.Cookie{
				Name:  "user_name",
				Value: username,
				Path:  "/",
			})
			
			// Store allowed modules
			var allowedModules string
			database.DB.QueryRow("SELECT COALESCE(allowed_modules, 'ac,work') FROM users WHERE id = ?", userID).Scan(&allowedModules)
			http.SetCookie(w, &http.Cookie{
				Name:  "user_modules",
				Value: allowedModules,
				Path:  "/",
			})
			
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	// Fallback to hardcoded admin for legacy/backup (Optional, maybe remove later)
	if username == "admin" && password == "admin123" {
		http.SetCookie(w, &http.Cookie{
			Name:  "session_token",
			Value: "authenticated",
			Path:  "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "user_role",
			Value: "master",
			Path:  "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "user_name",
			Value: "Admin Legacy",
			Path:  "/",
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/login.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Error": "Username atau password salah",
		"Year":  time.Now().Year(),
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "user_role",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "user_name",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil || cookie.Value != "authenticated" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func MasterMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user_role")
		if err != nil || cookie.Value != "master" {
			http.Error(w, "Unauthorized Access", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func ModuleMiddleware(module string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Master always has access
			if roleCookie, err := r.Cookie("user_role"); err == nil && roleCookie.Value == "master" {
				next(w, r)
				return
			}

			// Check if user has access to this module
			modulesCookie, err := r.Cookie("user_modules")
			if err != nil || !strings.Contains(modulesCookie.Value, module) {
				http.Error(w, "Akses Ditolak: Anda tidak memiliki izin untuk mengakses modul ini", http.StatusForbidden)
				return
			}
			next(w, r)
		}
	}
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, username, role, created_at, last_login, COALESCE(is_active, 1), COALESCE(allowed_modules, 'ac,work') FROM users ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt, &u.LastLogin, &u.IsActive, &u.AllowedModules); err != nil {
			log.Println("Error scanning user:", err)
			continue
		}
		users = append(users, u)
	}

	username := "Guest"
	role := ""
	if c, err := r.Cookie("user_name"); err == nil {
		username = c.Value
	}
	if c, err := r.Cookie("user_role"); err == nil {
		role = c.Value
	}

	tmpl, err := template.New("admin_users.html").Funcs(template.FuncMap{
		"contains": strings.Contains,
	}).ParseFiles("templates/admin_users.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Users":    users,
		"Year":     time.Now().Year(),
		"UserName": username,
		"UserRole": role,
	})
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")

	// Get allowed modules from checkboxes
	allowedModules := []string{}
	if r.FormValue("module_ac") == "on" {
		allowedModules = append(allowedModules, "ac")
	}
	if r.FormValue("module_work") == "on" {
		allowedModules = append(allowedModules, "work")
	}
	if role == "master" {
		allowedModules = []string{"ac", "work", "admin"} // Master gets all
	}
	modulesStr := strings.Join(allowedModules, ",")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Insert User
	res, err := database.DB.Exec("INSERT INTO users(username, password, role, allowed_modules, created_at) VALUES(?, ?, ?, ?, CURRENT_TIMESTAMP)", username, string(hashedPassword), role, modulesStr)
	if err != nil {
		log.Println("Error creating user:", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	userId, _ := res.LastInsertId()

	// Insert Permissions if role is user
	if role == "user" {
		var caps []string
		if r.FormValue("perm_ac_view") == "on" {
			caps = append(caps, "view")
		}
		if r.FormValue("perm_ac_create") == "on" {
			caps = append(caps, "create")
		}
		if r.FormValue("perm_ac_edit") == "on" {
			caps = append(caps, "edit")
		}
		if r.FormValue("perm_ac_delete") == "on" {
			caps = append(caps, "delete")
		}

		if len(caps) > 0 {
			capabilities := strings.Join(caps, ",")
			_, err = database.DB.Exec("INSERT INTO permissions(user_id, app_name, capabilities) VALUES(?, ?, ?)", userId, "ac", capabilities)
			if err != nil {
				log.Println("Error saving permissions:", err)
			}
		}
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func CheckPermission(r *http.Request, app, cap string) bool {
	cookie, err := r.Cookie("user_role")
	if err != nil {
		return false
	}
	if cookie.Value == "master" {
		return true
	}
	
	usernameCookie, err := r.Cookie("user_name")
	if err != nil {
		return false
	}
	username := usernameCookie.Value

	var capabilities string
	// Query by joining users and permissions
	err = database.DB.QueryRow(`
		SELECT p.capabilities 
		FROM permissions p 
		JOIN users u ON u.id = p.user_id 
		WHERE u.username = ? AND p.app_name = ?`, username, app).Scan(&capabilities)

	if err != nil {
		return false
	}

	return strings.Contains(capabilities, cap)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	// Prevent deleting self? (Optional check)

	_, err := database.DB.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func EditUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		// Update user modules
		allowedModules := []string{}
		if r.FormValue("module_ac") == "on" {
			allowedModules = append(allowedModules, "ac")
		}
		if r.FormValue("module_work") == "on" {
			allowedModules = append(allowedModules, "work")
		}
		
		// Check if user is master
		var role string
		database.DB.QueryRow("SELECT role FROM users WHERE id = ?", id).Scan(&role)
		if role == "master" {
			allowedModules = []string{"ac", "work", "admin"}
		}
		
		modulesStr := strings.Join(allowedModules, ",")
		_, err := database.DB.Exec("UPDATE users SET allowed_modules = ? WHERE id = ?", modulesStr, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	// GET: Show edit form
	var user models.User
	err := database.DB.QueryRow("SELECT id, username, role, COALESCE(allowed_modules, 'ac,work') FROM users WHERE id = ?", id).Scan(&user.ID, &user.Username, &user.Role, &user.AllowedModules)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	tmpl, err := template.New("edit_user.html").Funcs(template.FuncMap{
		"contains": strings.Contains,
	}).ParseFiles("templates/edit_user.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username := ""
	role := ""
	if c, err := r.Cookie("user_name"); err == nil {
		username = c.Value
	}
	if c, err := r.Cookie("user_role"); err == nil {
		role = c.Value
	}

	tmpl.Execute(w, map[string]interface{}{
		"User":     user,
		"Year":     time.Now().Year(),
		"UserName": username,
		"UserRole": role,
	})
}

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	username := ""
	if c, err := r.Cookie("user_name"); err == nil {
		username = c.Value
	}

	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user struct {
		ID        int
		Username  string
		Role      string
		LastLogin *time.Time
		CreatedAt time.Time
	}

	err := database.DB.QueryRow("SELECT id, username, role, last_login, created_at FROM users WHERE username = ?", username).Scan(
		&user.ID, &user.Username, &user.Role, &user.LastLogin, &user.CreatedAt,
	)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get recent login logs
	rows, _ := database.DB.Query("SELECT login_at, ip_address FROM login_logs WHERE user_id = ? ORDER BY login_at DESC LIMIT 5", user.ID)
	var loginLogs []struct {
		LoginAt   time.Time
		IPAddress string
	}
	if rows != nil {
		defer rows.Close()
		loc, _ := time.LoadLocation("Asia/Jakarta")
		for rows.Next() {
			var log struct {
				LoginAt   time.Time
				IPAddress string
			}
			rows.Scan(&log.LoginAt, &log.IPAddress)
			log.LoginAt = log.LoginAt.In(loc)
			loginLogs = append(loginLogs, log)
		}
	}

	tmpl, err := template.ParseFiles("templates/profile.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	role := ""
	if c, err := r.Cookie("user_role"); err == nil {
		role = c.Value
	}

	tmpl.Execute(w, map[string]interface{}{
		"User":      user,
		"LoginLogs": loginLogs,
		"Year":      time.Now().Year(),
		"UserName":  username,
		"UserRole":  role,
	})
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := ""
	if c, err := r.Cookie("user_name"); err == nil {
		username = c.Value
	}

	oldPassword := r.FormValue("old_password")
	newPassword := r.FormValue("new_password")

	if username == "" || oldPassword == "" || newPassword == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Verify old password
	var dbPassword string
	err := database.DB.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&dbPassword)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(oldPassword)) != nil {
		http.Error(w, "Old password incorrect", http.StatusUnauthorized)
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Update password
	_, err = database.DB.Exec("UPDATE users SET password = ? WHERE username = ?", string(hashedPassword), username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile?success=1", http.StatusSeeOther)
}

func ToggleUserStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	// Toggle is_active
	_, err := database.DB.Exec("UPDATE users SET is_active = NOT COALESCE(is_active, 1) WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}
