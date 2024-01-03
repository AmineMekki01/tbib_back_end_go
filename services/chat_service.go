package services

// ... other imports
import (
	"context"
	"fmt"
	"log"
	"net/http"
	"tbibi_back_end_go/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type CombinedUser struct {
	UserID string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
}


// GetChatsForUser retrieves all chat sessions for the specified user.
func GetChatsForUser(db *pgx.Conn, userID string) ([]models.Chat, error) {
	const query = `
	SELECT 
        c.id, 
        c.updated_at, 
        p.user_id, 
        COALESCE(pa.first_name, da.first_name) AS first_name, 
        COALESCE(pa.last_name, da.last_name) AS last_name
    FROM 
        chats AS c
    JOIN 
        participants AS p ON p.chat_id = c.id
    LEFT JOIN 
        patient_info AS pa ON pa.patient_id = p.user_id
    LEFT JOIN 
        doctor_info AS da ON da.doctor_id = p.user_id
    WHERE 
        p.user_id != $1
    AND
        p.deleted_at IS NULL
    AND
        c.id IN (SELECT chat_id FROM participants WHERE user_id = $1 AND deleted_at IS NULL)
    ;
    `
	rows, err := db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying chats for user %s: %v", userID, err)
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var chat models.Chat
		if err := rows.Scan(&chat.ID, &chat.UpdatedAt, &chat.UserID, &chat.FirstName, &chat.LastName); err != nil {
			return nil, fmt.Errorf("error scanning chat row: %v", err)
		}
        log.Println("chat: ", chat)
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chat rows: %v", err)
	}

	return chats, nil
}

func ListChatsForUser(db *pgx.Conn, c *gin.Context) {
	userID := c.Query("userID") 
    log.Println("userID: ", userID)
	chats, err := GetChatsForUser(db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chats"})
		return
	}
	c.JSON(http.StatusOK, chats)
}


type CombinedMessage struct {
    ChatID     string       `json:"chat_id"`
    SenderID   string       `json:"sender_id"`
    RecipientID string `json:"recipient_id"`
    Content    string    `json:"content"`
}

// SendMessage - sends a new message to a chat
func SendMessage(db *pgx.Conn, c *gin.Context) {
    var newMessage CombinedMessage
    if err := c.BindJSON(&newMessage); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message format: " + err.Error()})
        return
    }

    err := storeMessage(db, newMessage.SenderID, newMessage.ChatID, newMessage.Content)
    if err != nil {
        log.Printf("Failed to store message: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store message"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"status": "Message sent successfully"})
}

func storeMessage(db *pgx.Conn, senderID string, chatID string, content string) error {
    _, err := db.Exec(context.Background(),
        `INSERT INTO messages (chat_id, sender_id, content, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())`,
        chatID, senderID, content)
    return err
}

func SearchUsers(c *gin.Context, pool *pgxpool.Pool) {
    inputName := c.Param("username")
	
	var combinedUsers []CombinedUser

    queries := map[string]string{
        "patient": `SELECT patient_id, first_name, last_name FROM patient_info WHERE LOWER(first_name || ' ' || last_name) LIKE LOWER($1)`,
        "doctor":  `SELECT doctor_id, first_name, last_name FROM doctor_info WHERE LOWER(first_name || ' ' || last_name) LIKE LOWER($1)`,
    }

	for userType, query := range queries {
        rows, err := pool.Query(context.Background(), query, "%"+inputName+"%")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying " + userType + " table"})
            return
        }
        defer rows.Close()

        for rows.Next() {
            var user CombinedUser
            err := rows.Scan(&user.UserID, &user.FirstName, &user.LastName)
            if err != nil {
                continue  
            }
            combinedUsers = append(combinedUsers, user)
        }
    }
	
    c.JSON(http.StatusOK, gin.H{"users": combinedUsers})
}





func createChat(db *pgx.Conn, user1ID, user2ID int, user1Type, user2Type string) (int, error) {
    tx, err := db.Begin(context.Background())
    if err != nil {
        return 0, err
    }

    // Create a new chat
    var chatID int
    err = tx.QueryRow(context.Background(),
        `INSERT INTO chats (created_at, updated_at) VALUES (NOW(), NOW()) RETURNING id`).Scan(&chatID)
    if err != nil {
        tx.Rollback(context.Background())
        return 0, err
    }

    // Add participants
    _, err = tx.Exec(context.Background(),
        `INSERT INTO participants (chat_id, user_id, user_type, joined_at, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW(), NOW()), ($1, $4, $5, NOW(), NOW(), NOW())`,
        chatID, user1ID, user1Type, user2ID, user2Type)
    if err != nil {
        tx.Rollback(context.Background())
        return 0, err
    }

    err = tx.Commit(context.Background())
    if err != nil {
        return 0, err
    }

    return chatID, nil
}


func GetMessagesForChat(db *pgx.Conn, c *gin.Context) {
    chatID := c.Param("chatId")
    log.Println("chatID: ", chatID)
    rows, err := db.Query(context.Background(), `
    SELECT id, chat_id, sender_id, content, created_at, updated_at FROM messages 
    WHERE chat_id = $1 AND deleted_at IS NULL
    ORDER BY created_at ASC;`, chatID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
        return
    }
    defer rows.Close()

    var messages []models.Message
    for rows.Next() {
        var msg models.Message
        if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content, &msg.CreatedAt, &msg.UpdatedAt); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
            log.Println("Failed to retrieve messages : ", err)
            return
        }
        messages = append(messages, msg)
        log.Println("messages: ", messages)
    }
    if len(messages) == 0 {
        log.Println("No messages found for chat ID:", chatID)
    }
    c.JSON(http.StatusOK, gin.H{"messages": messages})
}


func findOrCreateChatWithUser(db *pgx.Conn, currentUserID, selectedUserID string) (string, error) {
    
    // Step 1: Check for an existing chat
    var chatID string
    err := db.QueryRow(context.Background(),
        `SELECT c.id FROM chats c
        JOIN participants p1 ON c.id = p1.chat_id
        JOIN participants p2 ON c.id = p2.chat_id
        WHERE p1.user_id = $1 AND p2.user_id = $2 AND p1.deleted_at IS NULL AND p2.deleted_at IS NULL AND p1.chat_id = p2.chat_id`,
        currentUserID, selectedUserID).Scan(&chatID)
    log.Println("chatID: ", chatID)
    if err == pgx.ErrNoRows {
        // Step 2: No existing chat, so create a new one
        tx, err := db.Begin(context.Background())
        if err != nil {
            log.Println("i am in err != nil {")
            log.Println("err : ", err)
            return "", err
        }
        log.Println("chat not found let's create one now")
        err = tx.QueryRow(context.Background(),
            `INSERT INTO chats (created_at, updated_at) VALUES (NOW(), NOW()) RETURNING id`).Scan(&chatID)
        if err != nil {
            log.Println("Ther was an error in creating a chat : ", err)
            tx.Rollback(context.Background())
            return "", err
        }

        // Add both users as participants
        _, err = tx.Exec(context.Background(),
            `INSERT INTO participants (chat_id, user_id, joined_at, created_at, updated_at) VALUES ($1, $2, NOW(), NOW(), NOW()), ($1, $3, NOW(), NOW(), NOW())`,
            chatID, currentUserID, selectedUserID)
        if err != nil {
            log.Println("Ther was an error in adding participants : ", err)
            tx.Rollback(context.Background())
            return "", err
        }

        err = tx.Commit(context.Background())
        if err != nil {
            return "", err
        }
    } else if err != nil {
        log.Printf("Error finding or creating chat: %v", err) // Detailed log
        return "", err
    }


    return chatID, nil
}

func FindOrCreateChatWithUser(db *pgx.Conn, c *gin.Context) {
    currentUserID := c.Query("currentUserId")
    selectedUserID := c.Query("selectedUserId")
    log.Println("currentUserId: ", currentUserID)
    log.Println("selectedUserId: ", selectedUserID)
    chatID, err := findOrCreateChatWithUser(db, currentUserID, selectedUserID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"chatId": chatID})
}