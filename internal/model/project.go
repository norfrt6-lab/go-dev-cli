package model

import "time"

type Project struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Path        string     `json:"path"`
	Language    string     `json:"language"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastOpened  *time.Time `json:"last_opened,omitempty"`
}
