package models

import "time"

type Task struct {
	ID            int
	Title         string
	Outcome       string
	Estimate      int       // in minutes
	Status        string    // inbox, todo, doing, done
	ScheduledDate time.Time // nullable
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CompletedAt   *time.Time
	AssignedTo    string     // Empty if unassigned
	DueDate       *time.Time // Nullable deadline
	Priority      string     // low, medium, high
	IsArchived    bool

	// Obstacle support
	Obstacles          []Obstacle
	OpenObstaclesCount int

	// Details support
	Details []TaskDetail
}

func (t Task) TotalDetails() int {
	return len(t.Details)
}

func (t Task) CompletedDetails() int {
	count := 0
	for _, d := range t.Details {
		if d.IsDone {
			count++
		}
	}
	return count
}

type TaskDetail struct {
	ID               int
	TaskID           int
	Description      string
	Progress         string
	Obstacle         string
	IsDone           bool
	ObstacleResolved bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
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
