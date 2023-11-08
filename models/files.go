package models

import "time"

type FolderInfo struct {
	ID        string    `json:"folder_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Type      string    `json:"file_type"`
	UserID    string    `json:"user_id"`
	UserType  string    `json:"user_type"`
	ParentID *string `json:"parent_id,omitempty"`
}