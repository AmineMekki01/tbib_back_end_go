package services

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"tbibi_back_end_go/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

func CreateFolder(c *gin.Context, pool *pgxpool.Pool) {
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
    if err != nil {
        log.Printf("Error reading body: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "can't read body"})
        return
    }

    bodyString := string(bodyBytes)
    log.Printf("Body received: %s", bodyString)

    c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

    var folderInfo models.FolderInfo
    if err := c.ShouldBind(&folderInfo); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
        return
    }
	
	folderPath := filepath.Join("./uploads", folderInfo.Name)
	
	log.Println(folderInfo)

	err = os.MkdirAll(folderPath, 0755) 
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}


	folderUUID, err := uuid.NewRandom()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate folder UUID"})
		return
	}
	folderInfo.ID = folderUUID.String()
	folderInfo.CreatedAt = time.Now()
	folderInfo.UpdatedAt = time.Now()

	conn, err := pool.Acquire(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not acquire database connection"})
		return
	}
	defer conn.Release()

	tx, err := conn.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not begin transaction"})
		return
	}

	_, err = tx.Exec(c.Request.Context(),
		"INSERT INTO folder_info (folder_id, name, created_at, updated_at, type, user_id, user_type) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		folderInfo.ID, folderInfo.Name, folderInfo.CreatedAt, folderInfo.UpdatedAt, folderInfo.Type, folderInfo.UserID, folderInfo.UserType)

	if err != nil {
		tx.Rollback(c.Request.Context())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert folder info"})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not commit transaction"})
		return
	}

	c.JSON(http.StatusOK, folderInfo)
}

func GetFolders(c *gin.Context, pool *pgxpool.Pool) {

	conn, err := pool.Acquire(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not acquire database connection"})
        return
    }
    defer conn.Release()

	userID := c.DefaultQuery("user_id", "")
    userType := c.DefaultQuery("user_type", "")
    fileType := c.DefaultQuery("file_type", "")

    rows, err := pool.Query(context.Background(), "SELECT folder_id, name, created_at, updated_at FROM folder_info WHERE user_id = $1 AND user_type = $2 AND type = $3", userID, userType, fileType)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var folders []models.FolderInfo
    for rows.Next() {
        var folder models.FolderInfo
        err := rows.Scan(&folder.ID, &folder.Name, &folder.CreatedAt, &folder.UpdatedAt)
				

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
		log.Println("yes")
        folders = append(folders, folder)
    }

    c.JSON(http.StatusOK, folders)
}
