package models

import (
	"time"
)

type Chat struct {
    ID                      string    `json:"id"`
    UserID             string `json:"user_id"`
    FirstName    string `json:"first_name"`
    LastName    string `json:"last_name"`
    UpdatedAt             time.Time `json:"updated_at"`

}

type Participant struct {
    ID                      string    `json:"id"`
    ChatID             string `json:"chat_id"`
    UserID    string `json:"user_id"`
    UserType    string `json:"user_type"`
    JoinedAt    string `json:"joined_at"`
    CreatedAt             string `json:"created_at"`
    UpdatedAt    string `json:"updated_at"`
    DeletedAt    string `json:"deleted_at"`
}

type Message struct {
    ID         string       `json:"id"`
    ChatID     string       `json:"chat_id"`
    SenderID   string       `json:"sender_id"`
    Content    string    `json:"content"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
    DeletedAt  time.Time `json:"deleted_at"`
}