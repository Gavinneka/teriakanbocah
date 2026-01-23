package main

import (
	"ac_tracker/internal/database"
	"fmt"

	_ "modernc.org/sqlite"
)

func main() {
	database.InitDB()
	rows, err := database.DB.Query("SELECT t.title, o.description FROM tasks t JOIN task_obstacles o ON t.id = o.task_id WHERE o.status = 'open'")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var t, o string
		rows.Scan(&t, &o)
		fmt.Printf("Task: %s | Obstacle: %s\n", t, o)
		count++
	}
	fmt.Printf("Total open obstacles: %d\n", count)
}
