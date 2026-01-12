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
}
