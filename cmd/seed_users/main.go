package main

import (
	"ac_tracker/internal/database"
	"database/sql"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func main() {
	database.InitDB()
	db := database.DB

	// Create Master User
	username := "master"
	password := "master123" // Change this in production!

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	var exists int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error checking user:", err)
	}

	if exists == 0 {
		_, err = db.Exec("INSERT INTO users(username, password, role) VALUES(?, ?, ?)", username, string(hashedPassword), "master")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Master user created successfully.")
	} else {
		fmt.Println("Master user already exists.")
	}

	// Create Staff User (System Kerja)
	seedUser(db, "staff", "staff123", "user", "work")

	// Create Technician User (AC Tracker)
	seedUser(db, "technician", "tech123", "user", "ac")
}

func seedUser(db *sql.DB, username, password, role, modules string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	var exists int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error checking user:", err)
		return
	}

	if exists == 0 {
		_, err = db.Exec("INSERT INTO users(username, password, role, allowed_modules) VALUES(?, ?, ?, ?)",
			username, string(hashedPassword), role, modules)
		if err != nil {
			log.Printf("Error creating %s: %v\n", username, err)
		} else {
			fmt.Printf("User %s created successfully.\n", username)
		}
	} else {
		fmt.Printf("User %s already exists.\n", username)
	}
}
