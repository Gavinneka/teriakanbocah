package models

import "time"

type Project struct {
	ID          int
	Name        string
	OutcomeGoal string
	Status      string
	CreatedAt   time.Time
}

type Task struct {
	ID            int
	Title         string
	Outcome       string
	Estimate      int       // in minutes
	Status        string    // inbox, todo, doing, done
	ProjectID     int       // nullable (0 if null)
	ScheduledDate time.Time // nullable
	ProjectName   string    // Joined field
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CompletedAt   *time.Time
	AssignedTo    string     // Empty if unassigned
	DueDate       *time.Time // Nullable deadline
	Priority      string // low, medium, high
	IsArchived    bool

	// Obstacle support
	Obstacles          []Obstacle
	OpenObstaclesCount int
}

type Obstacle struct {
	ID          int
	TaskID      int
	Description string
	Status      string // open, resolved
	CreatedAt   time.Time
	ResolvedAt  *time.Time
}

type ProblemLog struct {
	ID        int
	Problem   string
	Cause     string
	Solution  string
	CreatedAt time.Time
}

type Review struct {
	ID          int
	WeekStart   time.Time
	ReviewNotes string
	CreatedAt   time.Time
}
