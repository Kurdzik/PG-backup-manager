package handlers

import (
	"log"
	"net/http"
	"pg_bckup_mgr/auth"
	backup_manager "pg_bckup_mgr/backup-manager"
	"pg_bckup_mgr/db"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateBackupDestination(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("CreateBackupDestination handler called")
		var e db.Destination
		testFlag := c.Query("test_connection")
		log.Printf("CreateBackupDestination request with test_connection=%s", testFlag)
		if err := c.ShouldBindJSON(&e); err != nil {
			log.Printf("Error binding JSON in CreateBackupDestination: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}
		log.Printf("CreateBackupDestination request: Name=%s, ConnectionID=%d, EndpointURL=%s, BucketName=%s",
			e.Name, e.ConnectionID, e.EndpointURL, e.BucketName)
		encryptedAccessKeyID, err := auth.EncryptString(e.AccessKeyID)
		if err != nil {
			log.Printf("Error encrypting access key ID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to encrypt credentials",
				"error":   err.Error(),
			})
			return
		}
		e.AccessKeyID = encryptedAccessKeyID
		encryptedSecretAccessKey, err := auth.EncryptString(e.SecretAccessKey)
		if err != nil {
			log.Printf("Error encrypting secret access key: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to encrypt credentials",
				"error":   err.Error(),
			})
			return
		}
		e.SecretAccessKey = encryptedSecretAccessKey
		log.Printf("Credentials encrypted successfully for destination: %s", e.Name)
		if e.ConnectionID == 0 {
			log.Printf("Missing ConnectionID in CreateBackupDestination request")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Connection ID is required",
			})
			return
		}
		if e.Name == "" {
			log.Printf("Missing Name in CreateBackupDestination request")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Name is required",
			})
			return
		}
		if e.EndpointURL == "" {
			log.Printf("Missing EndpointURL in CreateBackupDestination request")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Endpoint URL is required",
			})
			return
		}
		if e.BucketName == "" {
			log.Printf("Missing BucketName in CreateBackupDestination request")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Bucket name is required",
			})
			return
		}
		if e.AccessKeyID == "" {
			log.Printf("Missing AccessKeyID in CreateBackupDestination request")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Access Key ID is required",
			})
			return
		}
		if e.SecretAccessKey == "" {
			log.Printf("Missing SecretAccessKey in CreateBackupDestination request")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Secret Access Key is required",
			})
			return
		}
		if testFlag == "true" {
			log.Printf("Testing S3 connection for destination: %s", e.Name)
			S3Client, err := backup_manager.NewS3Client(e.Name,
				e.EndpointURL,
				e.Region,
				e.BucketName,
				e.AccessKeyID,
				e.SecretAccessKey,
				e.UseSSL,
				e.VerifySSL,
			)
			if err != nil {
				log.Printf("Error creating S3 client for connection test: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  http.StatusInternalServerError,
					"message": "Failed to create S3 client",
					"error":   err.Error(),
				})
				return
			}
			if !S3Client.TestConnection() {
				log.Printf("S3 connection test failed for destination: %s", e.Name)
				c.JSON(http.StatusRequestTimeout, gin.H{
					"status":  http.StatusRequestTimeout,
					"message": "Could not reach specified backup destination",
				})
				return
			} else {
				log.Printf("S3 connection test successful for destination: %s", e.Name)
				c.JSON(http.StatusOK, gin.H{
					"status":  http.StatusOK,
					"message": "Conection Test successful",
				})
				return
			}
		}
		if err := conn.Create(&e).Error; err != nil {
			if isDuplicateKeyError(err) {
				log.Printf("Duplicate backup destination name attempted: %s", e.Name)
				c.JSON(http.StatusConflict, gin.H{
					"status":  http.StatusConflict,
					"message": "A backup destination with this name already exists",
				})
				return
			}
			if isForeignKeyError(err) {
				log.Printf("Invalid connection ID in CreateBackupDestination: %d", e.ConnectionID)
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  http.StatusBadRequest,
					"message": "Invalid connection ID - connection does not exist",
				})
				return
			}
			log.Printf("Error creating backup destination in database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to create backup destination",
				"error":   err.Error(),
			})
			return
		}
		log.Printf("Backup destination created successfully: %s (ID: %d)", e.Name, e.ID)
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Backup destination created successfully",
			"data":    e,
		})
	}
}

func ListAllBackupDestinations(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("ListAllBackupDestinations handler called")
		var destinations []db.Destination
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		connectionID := c.Query("connection_id")
		log.Printf("ListAllBackupDestinations request: page=%d, limit=%d, connection_id=%s", page, limit, connectionID)
		if page < 1 {
			page = 1
			log.Printf("Invalid page parameter, defaulting to 1")
		}
		if limit < 1 || limit > 100 {
			limit = 10
			log.Printf("Invalid limit parameter, defaulting to 10")
		}
		offset := (page - 1) * limit
		query := conn.Model(&db.Destination{})
		// Filter by connection_id if provided
		if connectionID != "" {
			if connID, err := strconv.ParseUint(connectionID, 10, 32); err == nil {
				query = query.Where("connection_id = ?", uint(connID))
				log.Printf("Filtering by connection_id: %d", uint(connID))
			} else {
				log.Printf("Invalid connection_id parameter: %s", connectionID)
			}
		}
		var total int64
		if err := query.Count(&total).Error; err != nil {
			log.Printf("Error counting backup destinations: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to count backup destinations",
				"error":   err.Error(),
			})
			return
		}
		log.Printf("Found %d total backup destinations", total)
		if err := query.Offset(offset).Limit(limit).Find(&destinations).Error; err != nil {
			log.Printf("Error retrieving backup destinations: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to retrieve backup destinations",
				"error":   err.Error(),
			})
			return
		}
		totalPages := (int(total) + limit - 1) / limit
		log.Printf("Retrieved %d backup destinations (page %d of %d)", len(destinations), page, totalPages)
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "OK",
			"data":    destinations,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
				"has_next":    page < totalPages,
				"has_prev":    page > 1,
			},
		})
	}
}

func UpdateBackupDestination(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("UpdateBackupDestination handler called")
		destinationId := c.Query("destination_id")
		log.Printf("UpdateBackupDestination request: destination_id=%s", destinationId)
		id, err := strconv.ParseUint(destinationId, 10, 32)
		if err != nil {
			log.Printf("Invalid destination ID format in UpdateBackupDestination: %s", destinationId)
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid ID format",
			})
			return
		}
		var existing db.Destination
		if err := conn.First(&existing, uint(id)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("Backup destination not found: ID=%d", uint(id))
				c.JSON(http.StatusNotFound, gin.H{
					"status":  http.StatusNotFound,
					"message": "Backup destination not found",
				})
				return
			}
			log.Printf("Error retrieving backup destination for update: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to retrieve backup destination",
				"error":   err.Error(),
			})
			return
		}
		log.Printf("Retrieved existing backup destination: %s (ID: %d)", existing.Name, existing.ID)
		var updates db.Destination
		if err := c.ShouldBindJSON(&updates); err != nil {
			log.Printf("Error binding JSON in UpdateBackupDestination: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}
		log.Printf("Update request parsed for destination: %s", existing.Name)
		if updates.ConnectionID != 0 && updates.ConnectionID != existing.ConnectionID {
			existing.ConnectionID = updates.ConnectionID
			log.Printf("Updated ConnectionID to: %d", updates.ConnectionID)
		}
		if updates.Name != "" && updates.Name != existing.Name {
			log.Printf("Updated Name from '%s' to '%s'", existing.Name, updates.Name)
			existing.Name = updates.Name
		}
		if updates.EndpointURL != "" {
			existing.EndpointURL = updates.EndpointURL
			log.Printf("Updated EndpointURL to: %s", updates.EndpointURL)
		}
		if updates.Region != "" {
			existing.Region = updates.Region
			log.Printf("Updated Region to: %s", updates.Region)
		}
		if updates.BucketName != "" {
			existing.BucketName = updates.BucketName
			log.Printf("Updated BucketName to: %s", updates.BucketName)
		}
		if updates.AccessKeyID != "" {
			existing.AccessKeyID = updates.AccessKeyID
			log.Printf("Updated AccessKeyID for destination: %s", existing.Name)
		}
		if updates.SecretAccessKey != "" {
			existing.SecretAccessKey = updates.SecretAccessKey
			log.Printf("Updated SecretAccessKey for destination: %s", existing.Name)
		}
		if updates.PathPrefix != existing.PathPrefix {
			existing.PathPrefix = updates.PathPrefix
			log.Printf("Updated PathPrefix to: %s", updates.PathPrefix)
		}
		existing.UseSSL = updates.UseSSL
		existing.VerifySSL = updates.VerifySSL
		log.Printf("Updated SSL settings: UseSSL=%t, VerifySSL=%t", updates.UseSSL, updates.VerifySSL)
		if err := conn.Save(&existing).Error; err != nil {
			if isDuplicateKeyError(err) {
				log.Printf("Duplicate backup destination name in update: %s", existing.Name)
				c.JSON(http.StatusConflict, gin.H{
					"status":  http.StatusConflict,
					"message": "A backup destination with this name already exists",
				})
				return
			}
			if isForeignKeyError(err) {
				log.Printf("Invalid connection ID in update: %d", existing.ConnectionID)
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  http.StatusBadRequest,
					"message": "Invalid connection ID - connection does not exist",
				})
				return
			}
			log.Printf("Error updating backup destination in database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to update backup destination",
				"error":   err.Error(),
			})
			return
		}
		log.Printf("Backup destination updated successfully: %s (ID: %d)", existing.Name, existing.ID)
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Backup destination updated successfully",
			"data":    existing,
		})
	}
}

func DeleteBackupDestination(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("DeleteBackupDestination handler called")
		destinationId := c.Query("destination_id")
		log.Printf("DeleteBackupDestination request: destination_id=%s", destinationId)
		id, err := strconv.ParseUint(destinationId, 10, 32)
		if err != nil {
			log.Printf("Invalid destination ID format in DeleteBackupDestination: %s", destinationId)
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid ID format",
			})
			return
		}
		var destination db.Destination
		if err := conn.First(&destination, uint(id)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("Backup destination not found for deletion: ID=%d", uint(id))
				c.JSON(http.StatusNotFound, gin.H{
					"status":  http.StatusNotFound,
					"message": "Backup destination not found",
				})
				return
			}
			log.Printf("Error retrieving backup destination for deletion: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to retrieve backup destination",
				"error":   err.Error(),
			})
			return
		}
		log.Printf("Retrieved backup destination for deletion: %s (ID: %d)", destination.Name, destination.ID)
		if err := conn.Delete(&destination).Error; err != nil {
			log.Printf("Error deleting backup destination from database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to delete backup destination",
				"error":   err.Error(),
			})
			return
		}
		log.Printf("Backup destination deleted successfully: %s (ID: %d)", destination.Name, destination.ID)
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Backup destination deleted successfully",
		})
	}
}

func isDuplicateKeyError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "duplicate") ||
		contains(errStr, "unique constraint") ||
		contains(errStr, "UNIQUE constraint failed")
}

func isForeignKeyError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "foreign key") ||
		contains(errStr, "violates foreign key constraint") ||
		contains(errStr, "FOREIGN KEY constraint failed")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOfSubstring(s, substr) >= 0)))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
