package main

import (
	"ac_tracker/internal/database"
	"ac_tracker/internal/handlers"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	database.InitDB()
	defer database.DB.Close()

	// Serve static files (if any needed in future, currently using CDN)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("GET /login", handlers.LoginPage)
	http.HandleFunc("POST /login", handlers.LoginHandler)
	http.HandleFunc("GET /logout", handlers.LogoutHandler)
	http.HandleFunc("GET /profile", handlers.AuthMiddleware(handlers.ProfilePage))
	http.HandleFunc("POST /profile/change-password", handlers.AuthMiddleware(handlers.ChangePassword))

	// Portal Home
	http.HandleFunc("GET /{$}", handlers.DashboardHandler)

	// AC App Routes (Module-protected)
	http.HandleFunc("GET /ac", handlers.AuthMiddleware(handlers.ModuleMiddleware("ac")(handlers.GetRecords)))
	http.HandleFunc("POST /ac/records", handlers.AuthMiddleware(handlers.ModuleMiddleware("ac")(handlers.CreateRecord)))
	http.HandleFunc("GET /ac/records/{id}/edit", handlers.AuthMiddleware(handlers.ModuleMiddleware("ac")(handlers.EditRecord)))
	http.HandleFunc("PUT /ac/records/{id}", handlers.AuthMiddleware(handlers.ModuleMiddleware("ac")(handlers.UpdateRecord)))
	http.HandleFunc("DELETE /ac/records/{id}", handlers.AuthMiddleware(handlers.ModuleMiddleware("ac")(handlers.DeleteRecord)))

	// Admin Routes (Master Only)
	http.HandleFunc("GET /admin/users", handlers.AuthMiddleware(handlers.MasterMiddleware(handlers.GetUsers)))
	http.HandleFunc("POST /admin/users/create", handlers.AuthMiddleware(handlers.MasterMiddleware(handlers.CreateUser)))
	http.HandleFunc("GET /admin/users/{id}/edit", handlers.AuthMiddleware(handlers.MasterMiddleware(handlers.EditUser)))
	http.HandleFunc("POST /admin/users/{id}/edit", handlers.AuthMiddleware(handlers.MasterMiddleware(handlers.EditUser)))
	http.HandleFunc("DELETE /admin/users/{id}", handlers.AuthMiddleware(handlers.MasterMiddleware(handlers.DeleteUser)))
	http.HandleFunc("POST /admin/users/{id}/toggle", handlers.AuthMiddleware(handlers.MasterMiddleware(handlers.ToggleUserStatus)))
	http.HandleFunc("POST /admin/users/{id}/reset_pass", handlers.AuthMiddleware(handlers.MasterMiddleware(handlers.AdminResetUserPassword)))

	// Work System Routes (Module-protected)
	http.HandleFunc("GET /work", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.WorkDashboard)))
	http.HandleFunc("POST /work/add", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.CreateSimpleTask)))
	http.HandleFunc("POST /work/toggle/{id}", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.ToggleTask)))
	http.HandleFunc("POST /work/notes/{id}", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.UpdateTaskNotes)))
	http.HandleFunc("GET /work/edit/{id}", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.EditSimpleTask)))
	http.HandleFunc("POST /work/edit/{id}", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.EditSimpleTask)))
	http.HandleFunc("POST /work/tasks/{id}/delete", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.DeleteTask)))
	http.HandleFunc("POST /work/tasks/{id}/details", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.AddTaskDetail)))
	http.HandleFunc("POST /work/tasks/{id}/details/{detailId}/edit", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.EditTaskDetail)))
	http.HandleFunc("POST /work/tasks/{id}/details/{detailId}/delete", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.DeleteTaskDetail)))
	http.HandleFunc("POST /work/tasks/{id}/details/{detailId}/toggle", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.ToggleTaskDetail)))
	http.HandleFunc("POST /work/tasks/{id}/details/{detailId}/toggle-obstacle", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.ToggleObstacle)))

	// Inbox and Task Drawer routes
	http.HandleFunc("GET /work/inbox", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.WorkInbox)))
	http.HandleFunc("POST /work/inbox", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.CreateInboxTask)))
	http.HandleFunc("GET /work/tasks/{id}/drawer", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.TaskDrawer)))
	http.HandleFunc("PUT /work/tasks/{id}", handlers.AuthMiddleware(handlers.ModuleMiddleware("work")(handlers.UpdateTaskDrawer)))

	// Improvement Log Routes
	http.HandleFunc("GET /improvement", handlers.AuthMiddleware(handlers.ImprovementList))
	http.HandleFunc("POST /improvement/add", handlers.AuthMiddleware(handlers.ImprovementAdd))
	http.HandleFunc("GET /improvement/edit/{id}", handlers.AuthMiddleware(handlers.ImprovementEdit))
	http.HandleFunc("POST /improvement/edit/{id}", handlers.AuthMiddleware(handlers.ImprovementEdit))
	http.HandleFunc("POST /improvement/delete/{id}", handlers.AuthMiddleware(handlers.ImprovementDelete))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
