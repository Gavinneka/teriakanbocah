package models

import "time"

type User struct {
	ID             int        `json:"id"`
	Username       string     `json:"username"`
	Password       string     `json:"password"` // Hashed
	Role           string     `json:"role"`     // 'master' or 'user'
	CreatedAt      time.Time  `json:"created_at"`
	LastLogin      *time.Time `json:"last_login"`      // Nullable
	IsActive       bool       `json:"is_active"`       // Default true
	AllowedModules string     `json:"allowed_modules"` // Comma-separated: ac,work,admin
}

type Permission struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	AppName      string `json:"app_name"`     // e.g., 'ac'
	Capabilities string `json:"capabilities"` // e.g., 'view,create,delete'
}
