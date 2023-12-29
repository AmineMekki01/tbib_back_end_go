package services

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
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
	// Parsing the form data
    var fileFolder models.FileFolder
    if err := c.ShouldBind(&fileFolder); err != nil {
        log.Println("Error parsing form:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
        return
    }
	
	// Generating a new UUID for the folder
	folderUUID, err := uuid.NewRandom()
	if err != nil {
        log.Println("Error generating folder UUID:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate folder UUID"})
		return
	}
	fileFolder.ID = folderUUID.String()
	fileFolder.CreatedAt = time.Now()
	fileFolder.UpdatedAt = time.Now()

    // set fileFolder.Size to NaN
    var size int64
    fileFolder.Size = size
    var ext *string
    fileFolder.Ext = ext

	// Create a new folder in the uploads directory
	folderPath := filepath.Join("./uploads", fileFolder.UserID)
	if fileFolder.ParentID != nil && *fileFolder.ParentID != "" {
        var parentFolderPath string
        if fileFolder.ParentID != nil {
            parentFolderPath, err = getParentFolderPath(*fileFolder.ParentID, pool)
        }
            if err != nil {
                log.Println("Error retrieving parent folder path:", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
            folderPath = filepath.Join(folderPath, parentFolderPath)
        }

    folderPath = filepath.Join(folderPath, fileFolder.Name)
    fileFolder.Path = folderPath
    err = os.MkdirAll(folderPath, 0755)
    if err != nil {
        log.Println("Error creating folder:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

	// Acquiring a connection from the connection pool
	conn, err := pool.Acquire(c.Request.Context())
	if err != nil {
        log.Println("Error acquiring connection:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not acquire database connection"})
		return
	}
	defer conn.Release()

	// Beginning a transaction
	tx, err := conn.Begin(c.Request.Context())
	if err != nil {
        log.Println("Error beginning transaction:", err )
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not begin transaction"})
		return
	}

    log.Println("hey")
	// Inserting the folder info into the database
	_, err = tx.Exec(c.Request.Context(),
    "INSERT INTO folder_file_info (id, name, created_at, updated_at, type, user_id, user_type, parent_id, size, extension, path) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
    fileFolder.ID, fileFolder.Name, fileFolder.CreatedAt, fileFolder.UpdatedAt, fileFolder.Type, fileFolder.UserID, fileFolder.UserType, fileFolder.ParentID, fileFolder.Size, fileFolder.Ext, fileFolder.Path)
	if err != nil {
        log.Println("Error inserting folder info:", err )
		tx.Rollback(c.Request.Context())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert folder info"})
		return
	}

	// Committing the transaction
	if err := tx.Commit(c.Request.Context()); err != nil {
        log.Println("Error committing transaction:", err )
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not commit transaction"})
		return
	}

	// Returning the folder info as a JSON response
	c.JSON(http.StatusOK, fileFolder)
}

func GetBreadcrumbs(c *gin.Context, pool *pgxpool.Pool) {
    folderID := c.Param("folderId")
	// Acquiring a connection from the connection pool
    breadcrumbs, err := getParentFolders(folderID, pool)
    if err != nil {
        log.Println("Error getting parent folders:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, breadcrumbs)
}

func getParentFolders(folderID string, pool *pgxpool.Pool) ([]models.FileFolder, error) {
    var breadcrumbs []models.FileFolder
	// getting the parent folder
    for folderID != "" {
        var folder models.FileFolder
        row := pool.QueryRow(context.Background(), "SELECT id, name, parent_id FROM folder_file_info WHERE id = $1", folderID)
        err := row.Scan(&folder.ID, &folder.Name, &folder.ParentID)
        if err != nil {
            log.Println("Error getting parent folder:", err)
            return nil, err 
        }
        breadcrumbs = append([]models.FileFolder{folder}, breadcrumbs...) 
        folderID = ""
        if folder.ParentID != nil {
            folderID = *folder.ParentID
        }
    }
    return breadcrumbs, nil
}

func getParentFolderPath(folderID string, pool *pgxpool.Pool) (string, error) {
	// getting the parent folder path
    log.Println("Got the folderID in the getParentFolderPath function :", folderID)
	var parentFolder models.FileFolder
	row := pool.QueryRow(context.Background(), "SELECT name, parent_id FROM folder_file_info WHERE id = $1", folderID)
	err := row.Scan(&parentFolder.Name, &parentFolder.ParentID)
	if err != nil {
        log.Println("Error getting parent folder:", err)
		return "", err 
	}
	if parentFolder.ParentID == nil {
		return parentFolder.Name, nil
	}
	parentPath, err := getParentFolderPath(*parentFolder.ParentID, pool)
	if err != nil {
        log.Println("Error getting parent folder path:", err)
		return "", err 
	}
	return filepath.Join(parentPath, parentFolder.Name), nil
}

func GetFolders(c *gin.Context, pool *pgxpool.Pool) {

	// Acquiring a connection from the connection pool
	conn, err := pool.Acquire(c.Request.Context())
    if err != nil {
        log.Println("Error acquiring connection:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not acquire database connection"})
        return
    }
    defer conn.Release()

	// Extracting the query parameters
	userID := c.DefaultQuery("user_id", "")
    userType := c.DefaultQuery("user_type", "")
	parentID := c.Query("parent_id")


	// Preparing the base query
	baseQuery := "SELECT id, name, created_at, updated_at, type, extension, path FROM folder_file_info WHERE user_id = $1 AND user_type = $2"
    args := []interface{}{userID, userType}

	// Adding the parent_id condition if it is specified
	if parentID != "" {
        // Fetching subfolders for a given parent_id
        baseQuery += " AND parent_id = $3"
        args = append(args, parentID)
    } else {
        // Fetching root folders (no parent_id)
        baseQuery += " AND parent_id IS NULL"
    }

    // Executing the query with the prepared arguments
    rows, err := conn.Query(c.Request.Context(), baseQuery, args...)
    if err != nil {
        log.Println("Error executing query:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

	// Initializing a slice to hold the retrieved folders
    var folders []models.FileFolder

	// Iterating over the rows in the result set
    for rows.Next() {
        var folder models.FileFolder
        var path *string
        err := rows.Scan(&folder.ID, &folder.Name, &folder.CreatedAt, &folder.UpdatedAt, &folder.Type, &folder.Ext, &path)
		
		if err != nil {
            log.Println("Error scanning row:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        if path != nil {
            folder.Path = *path
        } else {
            folder.Path = "" 
        }
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
        log.Println("Error acquiring connection:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not acquire database connection"})
        return
    }
    defer conn.Release()

    // Preparing the SQL query to select folders with the specified parent_id
    query := "SELECT id, name, created_at, updated_at FROM folder_file_info WHERE parent_id = $1"

    // Executing the query with the parentID as the parameter
    rows, err := conn.Query(c.Request.Context(), query, parentID)
    if err != nil {
        log.Println("Error executing query:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    // Initializing a slice to hold the retrieved folders
    var subfolders []models.FileFolder

    // Iterating over the rows in the result set
    for rows.Next() {
        var folder models.FileFolder
        err := rows.Scan(&folder.ID, &folder.Name, &folder.CreatedAt, &folder.UpdatedAt)
        if err != nil {
            log.Println("Error scanning row:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        subfolders = append(subfolders, folder)
    }
    if err = rows.Err(); err != nil {
        log.Println("Error scanning row:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, subfolders)
}


func DeleteFolderAndContents(c *gin.Context, pool *pgxpool.Pool) {
    // Parse the JSON body to get the folder ID to delete

    var request struct {
        FolderID string `json:"folderId"`
    }
    if err := c.ShouldBindJSON(&request); err != nil {
        log.Println("Error parsing JSON:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    // Start a transaction
    tx, err := pool.Begin(c.Request.Context())
    if err != nil {
        log.Println("Error beginning transaction:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not begin transaction"})
        return
    }
    defer tx.Rollback(c.Request.Context())

    // Use a CTE to recursively get all file and folder IDs within the target folder
    cteQuery := `
        WITH RECURSIVE subfolders AS (
            SELECT id FROM folder_file_info WHERE id = $1
            UNION ALL
            SELECT fi.id FROM folder_file_info fi
            INNER JOIN subfolders s ON s.id = fi.parent_id
        )
        DELETE FROM folder_file_info WHERE id IN (SELECT id FROM subfolders);
        `
    // Execute the CTE query to delete all subfolders and files in the database
    if _, err := tx.Exec(c.Request.Context(), cteQuery, request.FolderID); err != nil {
        log.Println("Error deleting folder contents:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete folder contents"})
        return
    }

    // Commit the transaction - This is missing from the code snippet provided
    if err := tx.Commit(c.Request.Context()); err != nil {
        log.Println("Error committing transaction:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not commit transaction"})
        return
    }

    // Commit the transaction
    _, err = tx.Exec(c.Request.Context(), "DELETE FROM folder_file_info WHERE id = $1", request.FolderID)
    if err != nil {
        log.Println("Error deleting root folder:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete root folder"})
        return
    }

    // Delete the folder from the filesystem
    folderPath := filepath.Join("./uploads", request.FolderID)
    if err := os.RemoveAll(folderPath); err != nil {
        log.Printf("Error deleting folder from filesystem: %s\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete folder from filesystem"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Folder and contents deleted successfully"})
}

func UpdateFolderName(c *gin.Context, pool *pgxpool.Pool) {
    folderID := c.Param("folderId")
    var updateRequest struct {
        Name string `json:"name"`
    }

    if err := c.ShouldBindJSON(&updateRequest); err != nil {
        log.Println("Error parsing JSON:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    conn, err := pool.Acquire(c.Request.Context())
    if err != nil {
        log.Println("Error acquiring connection:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not acquire database connection"})
        return
    }
    defer conn.Release()

    // use the folderID to update the folder name
    
    _, err = conn.Exec(c.Request.Context(), "UPDATE folder_file_info SET name = $1, updated_at = $2 WHERE id = $3",
        updateRequest.Name, time.Now(), folderID)

    if err != nil {
        log.Println("Error updating folder name:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update folder name"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Folder name updated successfully"})
}

func generateFilePath(userID string, parentFolderId string, filename string, pool *pgxpool.Pool) (string, error) {
	var folderPath string
	var err error

	log.Println("parentFolderId : ", parentFolderId)
	if parentFolderId != "" {
		folderPath, err = getParentFolderPath(parentFolderId, pool)
        log.Println("the folder path in from generate the parent folder id : ", folderPath)
		if err != nil {
            log.Println("Error retrieving parent folder path:", err)
			return "", err
		}
	} else {
		folderPath = filepath.Join("./uploads",  )
	}

	err = os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
        log.Println("Error creating folder:", err)
		return "", err
	}

    if !filepath.HasPrefix(folderPath, userID) {
        if !filepath.HasPrefix(folderPath, "uploads") {
            folderPath = filepath.Join("uploads", userID, folderPath)
        } else {
            log.Println("Got the folderPath in the generateFilePath function :", folderPath)

            parts := filepath.SplitList("uploads")
            if len(parts) > 1 {
                folderPath = filepath.Join(parts[0], userID, folderPath)
            } else {
                folderPath = filepath.Join(parts[0], userID)
            }
        }
    }

	fullFilePath := filepath.Join(folderPath, filename)

    _, err = os.Stat(fullFilePath)
    if err == nil {
        log.Println("File already exists")
        return "", fmt.Errorf("file already exists")
    }
	return fullFilePath, nil
}


func UploadFile(c *gin.Context, pool *pgxpool.Pool) {

    var fileInfo models.FileFolder

    err := c.Request.ParseMultipartForm(10 << 20) // 10 MB
    if err != nil {
        log.Println("Error parsing multipart form:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse multipart form"})
        return
    }

    file, handler, err := c.Request.FormFile("file")
    if err != nil {
        log.Println("Error retrieving file from request:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Could not get file from request"})
        return
    }
    defer file.Close()

    if parentFolderID := c.Request.FormValue("parentFolderId");
    parentFolderID != "" {
        if _, err := uuid.Parse(parentFolderID); err != nil {
            log.Printf("Invalid parentFolderId: %s\n", err)
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parentFolderId"})
            return
        }
        fileInfo.ParentID = &parentFolderID
    } else {
        fileInfo.ParentID = nil
    }

    // Generate the file info
    fileInfo.CreatedAt = time.Now()
    fileInfo.UpdatedAt = time.Now()
    fileInfo.Type = c.Request.FormValue("fileType")
    ext := c.Request.FormValue("fileExt")
    fileInfo.Ext = &ext
    fileInfo.UserID = c.Request.FormValue("userId")
    fileInfo.UserType = c.Request.FormValue("userType")
    fileInfo.Size = handler.Size
    fileInfo.Name = handler.Filename
    id, _ := uuid.NewRandom()
    fileInfo.ID = id.String()

    // Generate the file path
    var parentID string
    if fileInfo.ParentID != nil {
        parentID = *fileInfo.ParentID
    }
    var filePath string
    filePath, err = generateFilePath(fileInfo.UserID, parentID, fileInfo.Name, pool)
    if err != nil {
        log.Printf("Error generating file path: %s\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate file path"})
        return
    }

    fileInfo.Path = filePath    
    // save the file to the filesystem
    newFile, err := os.Create(filePath)
    if err != nil {
        log.Printf("Error creating file: %s\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
        return
    }

    _, err = file.Seek(0, 0)
    if err != nil {
        log.Printf("Error seeking to beginning of file: %s\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seek to beginning of file"})
        return
    }

    _, err = newFile.Seek(0, 0)
    if err != nil {
        log.Printf("Error seeking to beginning of new file: %s\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seek to beginning of new file"})
        return
    }

    _, err = io.Copy(newFile, file)
    if err != nil {
        log.Printf("Error copying file data: %s\n", err)    
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy file data"})
        return
    }

    // Close the file
    if err = newFile.Close(); err != nil {
        log.Printf("Error closing file: %s\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close file"})
        return
    }

    // Insert the file info into the database
    _, err = pool.Exec(c.Request.Context(), "INSERT INTO folder_file_info (id, name, created_at, updated_at, type, size, extension, user_id, user_type, parent_id, path) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
        fileInfo.ID, fileInfo.Name, fileInfo.CreatedAt, fileInfo.UpdatedAt, fileInfo.Type, fileInfo.Size, fileInfo.Ext, fileInfo.UserID, fileInfo.UserType, fileInfo.ParentID, fileInfo.Path)
    if err != nil { 
        log.Printf("Error inserting file info: %s\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert file info"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}


func DownloadFile(c *gin.Context, pool *pgxpool.Pool) {
    log.Println("DownloadFile function called")

    fileId := c.Param("fileId")
    log.Printf("Requested file ID: %s", fileId)

    // Acquire a connection from the pool
    conn, err := pool.Acquire(c.Request.Context())
    if err != nil {
        log.Printf("Error acquiring connection: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not acquire database connection"})
        return
    }
    defer conn.Release()

    // Retrieve file information
    var file models.FileFolder
    err = conn.QueryRow(c.Request.Context(), "SELECT id, name, path FROM folder_file_info WHERE id = $1", fileId).Scan(&file.ID, &file.Name, &file.Path)
    if err != nil {
        log.Printf("Error retrieving file information: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve file information"})
        return
    }

    log.Printf("File Type: %s, Path: %s", file.Type, file.Path)

    if file.Type == "folder" {
        log.Println("Processing as a folder")
        zipFilePath, err := createZipFromFolder(file.Path)
        if err != nil {
            log.Printf("Error creating zip file: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create zip file"})
            return
        }
        log.Printf("Zip file created: %s", zipFilePath)    
        c.File(zipFilePath)
        // Optionally delete the zip file after serving it
    } else {
        log.Println("Processing as a regular file")
        // It's a file, serve it as before
        c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
        c.File(file.Path)
    }
}

func createZipFromFolder(folderPath string) (string, error) {
    log.Printf("Creating ZIP from folder: %s", folderPath)
    zipFileName := "tempFolderZip.zip" 
    zipFilePath := filepath.Join("./uploads", zipFileName)
    log.Printf("Zip file path: %s", zipFilePath)

    newZipFile, err := os.Create(zipFilePath)
    if err != nil {
        log.Printf("Error creating zip file: %v", err)
        return "", err
    }
    defer newZipFile.Close()

    zipWriter := zip.NewWriter(newZipFile)
    defer zipWriter.Close()

    // Function to recursively add files and folders to the zip
    err = addFilesToZip(zipWriter, folderPath, "")
    if err != nil {
        log.Printf("Error adding files to zip: %v", err)
        return "", err
    }

    return zipFilePath, nil
}

func addFilesToZip(zipWriter *zip.Writer, basePath, baseInZip string) error {
    log.Printf("Adding files to ZIP: BasePath: %s, BaseInZip: %s", basePath, baseInZip)
    files, err := ioutil.ReadDir(basePath)
    if err != nil {
        log.Printf("Error reading directory: %v", err)
        return err
    }

    for _, file := range files {
        currentPath := filepath.Join(basePath, file.Name())
        log.Printf("Processing file: %s", currentPath)

        // Check if it's a directory, recursively add its files
        if file.IsDir() {
            newBaseInZip := filepath.Join(baseInZip, file.Name())
            err = addFilesToZip(zipWriter, currentPath, newBaseInZip)
            if err != nil {
                log.Printf("Error adding directory to zip: %v", err)
                return err
            }
        } else {
            // It's a file, add it to the zip
            data, err := ioutil.ReadFile(currentPath)
            if err != nil {
                log.Printf("Error reading file: %v", err)
                return err
            }

            f, err := zipWriter.Create(filepath.Join(baseInZip, file.Name()))
            if err != nil {
                log.Printf("Error creating zip entry: %v", err)
                return err
            }
            _, err = f.Write(data)
            if err != nil {
                log.Printf("Error writing to zip entry: %v", err)
                return err
            }
        }
    }

    return nil
}


