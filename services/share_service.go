// File: services/share.go

package services

import (
	"context"
	"log"
	"net/http"
	"tbibi_back_end_go/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)


type ShareRequest struct {
    SharedWithID string   `json:"sharedWithID"`
    ItemIDs    []string `json:"itemIDs"`
    UserID     string   `json:"userID"`
    UserType     string   `json:"userType"`
}
// i already have a GetAllDoctors function in doctor_service.go but this one is going to be different in the futur. it will be used to get all the doctors that the patient has followed (or have gave him access to his files. don't know yet which one i will sue) I will include getting the doctors photo as well.
func ListDoctors(c *gin.Context, db *pgxpool.Pool) {
    rows, err := db.Query(context.Background(), "SELECT doctor_id, first_name , last_name, specialty FROM users WHERE user_type = 'doctor'")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve doctors list"})
        return
    }

    defer rows.Close()

    var doctors []models.Doctor 
    for rows.Next() {
        var doctor models.Doctor
        if err := rows.Scan(&doctor.DoctorID, &doctor.FirstName, &doctor.LastName, &doctor.Specialty); err != nil {
            continue 
        }
        doctors = append(doctors, doctor)
    }

    c.JSON(http.StatusOK, doctors)
}

func ShareItem(c *gin.Context, db *pgxpool.Pool) {
    var req ShareRequest

    // Bind JSON body to the ShareRequest struct
    if err := c.BindJSON(&req); err != nil {
        log.Printf("Error binding JSON: %v\n", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

	// if req.UserType == "patient" && req.RecipientType != "doctor" {
    //     c.JSON(http.StatusBadRequest, gin.H{"error": "Patients can only share with doctors"})
    //     return
    // } else if req.UserType == "doctor" && req.RecipientType != "doctor" {
    //     c.JSON(http.StatusBadRequest, gin.H{"error": "Doctors can only share with other doctors"})
    //     return
    // }
    // Iterate over each itemID and share it with the specified user
    for _, itemID := range req.ItemIDs {
        sharedItem := models.SharedItem{
            ItemID:    itemID,
            SharedBy:  req.UserID,
            SharedWith: req.SharedWithID,
            SharedAt:  time.Now(),
        }

        // Prepare SQL query to insert the new shared item record
        sql := `INSERT INTO shared_items (item_id, shared_by_id, shared_with_id, shared_at) VALUES ($1, $2, $3, $4)`
        _, err := db.Exec(context.Background(), sql, sharedItem.ItemID, sharedItem.SharedBy, sharedItem.SharedWith, sharedItem.SharedAt) 
        if err != nil {  
            log.Printf("Unable to insert the new shared item record into the database: %v\n", err)
            // Continue attempting to share the remaining items even if one fails
            continue
        }
    }

    // Respond with success message
    c.JSON(http.StatusOK, gin.H{"message": "Items shared successfully"})
}

// Retrieve items shared with the user
func GetSharedWithMe(c *gin.Context, db *pgxpool.Pool) {
	userID := c.Query("userId")
	var items []models.FileFolder
	log.Println("userID", userID)
	// SQL query to retrieve the items shared with the user
	sql := `SELECT 
    f.id, 
    f.name, 
    f.created_at, 
    f.updated_at, 
    f.type, 
    f.size, 
    f.extension, 
    f.user_id, 
    f.user_type, 
    f.parent_id, 
    f.path 
	FROM shared_items s 
	JOIN folder_file_info f ON s.item_id = f.id 
	WHERE s.shared_with_id = $1`

	rows, err := db.Query(context.Background(), sql, userID)	
	if err != nil {	
		log.Printf("Unable to execute the select query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve shared items"})
		return 
	}

	defer rows.Close()	

	for rows.Next() {	
		var item models.FileFolder	
		if err := rows.Scan(	
			&item.ID, 	
			&item.Name, 	
			&item.CreatedAt, 	
			&item.UpdatedAt,	
			&item.Type,	
			&item.Size,	
			&item.Ext,	
			&item.UserID,	
			&item.UserType,		
			&item.ParentID,		
			&item.Path,	
			); err != nil {	
			log.Printf("Unable to scan the row. %v\n", err)	
			return 
		}	
		items = append(items, item)	
	}

	if len(items) == 0 {
		// If no items, return an empty array instead of null
		c.JSON(http.StatusOK, []models.FileFolder{})
	  } else {
		c.JSON(http.StatusOK, items)
	  }
	}

// Retrieve items shared by the user
func GetSharedByMe(c *gin.Context, db *pgxpool.Pool) {
	userID := c.Query("userId")
	var items []models.FileFolder
	log.Println("userID", userID)
	// SQL query to retrieve the items shared with the user
	sql := `SELECT 
    f.id, 
    f.name, 
    f.created_at, 
    f.updated_at, 
    f.type, 
    f.size, 
    f.extension, 
    f.user_id, 
    f.user_type, 
    f.parent_id, 
    f.path 
	FROM shared_items s 
	JOIN folder_file_info f ON s.item_id = f.id 
	WHERE s.shared_by_id = $1`

	rows, err := db.Query(context.Background(), sql, userID)	
	if err != nil {	
		log.Printf("Unable to execute the select query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve shared items"})
		return 
	}

	defer rows.Close()	

	for rows.Next() {	
		var item models.FileFolder	
		log.Println("rows", rows)
		if err := rows.Scan(	
			&item.ID, 	
			&item.Name, 	
			&item.CreatedAt, 	
			&item.UpdatedAt,	
			&item.Type,	
			&item.Size,	
			&item.Ext,	
			&item.UserID,	
			&item.UserType,		
			&item.ParentID,		
			&item.Path,	
			); err != nil {	
			log.Printf("Unable to scan the row. %v\n", err)	
			return 
		}	
		items = append(items, item)	
	}
	if len(items) == 0 {
    // If no items, return an empty array instead of null
    c.JSON(http.StatusOK, []models.FileFolder{})
  } else {
    c.JSON(http.StatusOK, items)
  }
}