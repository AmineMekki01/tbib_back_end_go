package models

import "time"

type FileFolder struct {
	ID        string    `json:"folder_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Type      string    `json:"file_type"`
	Size 	int64     `json:"size"`
	Ext       *string    `json:"extension"`
	UserID    string    `json:"user_id"`
	UserType  string    `json:"user_type"`
	ParentID *string `json:"parent_id,omitempty"`
	Path 	string    `json:"path"`
}
