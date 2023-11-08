package services

import (
	"context"
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
	// Parsing the form data
    var folderInfo models.FolderInfo
    if err := c.ShouldBind(&folderInfo); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
        return
    }
	
	// Generating a new UUID for the folder
	folderUUID, err := uuid.NewRandom()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate folder UUID"})
		return
	}
	folderInfo.ID = folderUUID.String()
	folderInfo.CreatedAt = time.Now()
	folderInfo.UpdatedAt = time.Now()


	// Create a new folder in the uploads directory
	folderPath := filepath.Join("./uploads", folderInfo.UserID)
	if folderInfo.ParentID != nil && *folderInfo.ParentID != "" {
	var parentFolderPath string
	if folderInfo.ParentID != nil {
		parentFolderPath, err = getParentFolderPath(*folderInfo.ParentID, pool)
	}
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        folderPath = filepath.Join(folderPath, parentFolderPath)
    }
    folderPath = filepath.Join(folderPath, folderInfo.Name)
    err = os.MkdirAll(folderPath, 0755)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

	// Acquiring a connection from the connection pool
	conn, err := pool.Acquire(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not acquire database connection"})
		return
	}
	defer conn.Release()

	// Beginning a transaction
	tx, err := conn.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not begin transaction"})
		return
	}

	// Inserting the folder info into the database
	_, err = tx.Exec(c.Request.Context(),
    "INSERT INTO folder_info (folder_id, name, created_at, updated_at, type, user_id, user_type, parent_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
    folderInfo.ID, folderInfo.Name, folderInfo.CreatedAt, folderInfo.UpdatedAt, folderInfo.Type, folderInfo.UserID, folderInfo.UserType, folderInfo.ParentID)
	if err != nil {
		tx.Rollback(c.Request.Context())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert folder info"})
		return
	}

	// Committing the transaction
	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not commit transaction"})
		return
	}

	// Returning the folder info as a JSON response
	c.JSON(http.StatusOK, folderInfo)
}

func GetBreadcrumbs(c *gin.Context, pool *pgxpool.Pool) {
    folderID := c.Param("folderId")
	// Acquiring a connection from the connection pool
    breadcrumbs, err := getParentFolders(folderID, pool)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, breadcrumbs)
}

func getParentFolders(folderID string, pool *pgxpool.Pool) ([]models.FolderInfo, error) {
    var breadcrumbs []models.FolderInfo
	// getting the parent folder
    for folderID != "" {
        var folder models.FolderInfo
        row := pool.QueryRow(context.Background(), "SELECT folder_id, name, parent_id FROM folder_info WHERE folder_id = $1", folderID)
        err := row.Scan(&folder.ID, &folder.Name, &folder.ParentID)
        if err != nil {
            return nil, err 
        }
        breadcrumbs = append([]models.FolderInfo{folder}, breadcrumbs...) 
        folderID = ""
        if folder.ParentID != nil {
            folderID = *folder.ParentID
        }
    }
    return breadcrumbs, nil
}

func getParentFolderPath(folderID string, pool *pgxpool.Pool) (string, error) {
	// getting the parent folder path
	var parentFolder models.FolderInfo
	row := pool.QueryRow(context.Background(), "SELECT name, parent_id FROM folder_info WHERE folder_id = $1", folderID)
	err := row.Scan(&parentFolder.Name, &parentFolder.ParentID)
	if err != nil {
		return "", err 
	}

	if parentFolder.ParentID == nil {
		return parentFolder.Name, nil
	}
	parentPath, err := getParentFolderPath(*parentFolder.ParentID, pool)
	if err != nil {
		return "", err 
	}
	return filepath.Join(parentPath, parentFolder.Name), nil
}

func GetFolders(c *gin.Context, pool *pgxpool.Pool) {

	// Acquiring a connection from the connection pool
	conn, err := pool.Acquire(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not acquire database connection"})
        return
    }
    defer conn.Release()

	// Extracting the query parameters
	userID := c.DefaultQuery("user_id", "")
    userType := c.DefaultQuery("user_type", "")
    fileType := c.DefaultQuery("file_type", "")
	parentID := c.Query("parent_id")


	// Preparing the base query
	baseQuery := "SELECT folder_id, name, created_at, updated_at FROM folder_info WHERE user_id = $1 AND user_type = $2 AND type = $3"
    args := []interface{}{userID, userType, fileType}

	// Adding the parent_id condition if it is specified
	if parentID != "" {
        // Fetching subfolders for a given parent_id
        baseQuery += " AND parent_id = $4"
        args = append(args, parentID)
    } else {
        // Fetching root folders (no parent_id)
        baseQuery += " AND parent_id IS NULL"
    }

    // Executing the query with the prepared arguments
    rows, err := conn.Query(c.Request.Context(), baseQuery, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

	// Initializing a slice to hold the retrieved folders
    var folders []models.FolderInfo

	// Iterating over the rows in the result set
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



func GetSubfolders(c *gin.Context, pool *pgxpool.Pool) {
	// Extracting the parent_id from the URL parameter
	parentID := c.Param("folderId")

    // Acquiring a connection from the connection pool
    conn, err := pool.Acquire(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not acquire database connection"})
        return
    }
    defer conn.Release()

    // Preparing the SQL query to select folders with the specified parent_id
    query := "SELECT folder_id, name, created_at, updated_at FROM folder_info WHERE parent_id = $1"

    // Executing the query with the parentID as the parameter
    rows, err := conn.Query(c.Request.Context(), query, parentID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    // Initializing a slice to hold the retrieved folders
    var subfolders []models.FolderInfo

    // Iterating over the rows in the result set
    for rows.Next() {
        var folder models.FolderInfo
        err := rows.Scan(&folder.ID, &folder.Name, &folder.CreatedAt, &folder.UpdatedAt)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        subfolders = append(subfolders, folder)
    }
    if err = rows.Err(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, subfolders)
}