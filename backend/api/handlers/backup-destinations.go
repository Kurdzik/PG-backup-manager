package handlers

import (
	"log"
	"net/http"
	bck "pg_bckup_mgr/backup-manager"
	"pg_bckup_mgr/db"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateBackupDestination(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var e db.Destination

		testFlag := c.Query("test_connection")

		if err := c.ShouldBindJSON(&e); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Invalid request body",
				"error":  err.Error(),
			})
			return
		}

		if e.ConnectionID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Connection ID is required",
			})
			return
		}

		if e.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Name is required",
			})
			return
		}

		if e.EndpointURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Endpoint URL is required",
			})
			return
		}

		if e.BucketName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bucket name is required",
			})
			return
		}

		if e.AccessKeyID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Access Key ID is required",
			})
			return
		}

		if e.SecretAccessKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Secret Access Key is required",
			})
			return
		}

		if testFlag == "true" {
			S3Client, err := bck.NewS3Client(e.Name,
				e.EndpointURL,
				e.Region,
				e.BucketName,
				e.AccessKeyID,
				e.SecretAccessKey,
				e.UseSSL,
				e.VerifySSL,
			)
			if err != nil {
				log.Printf("Error creating S3 client: %v", err)
			}

			if !S3Client.TestConnection() {
				c.JSON(http.StatusRequestTimeout, gin.H{
					"status": "Could not reach specified backup destination",
				})
				return
			} else {
				c.JSON(http.StatusOK, gin.H{
					"status": "Conection Test successful",
				})
				return
			}
		}

		if err := conn.Create(&e).Error; err != nil {
			if isDuplicateKeyError(err) {
				c.JSON(http.StatusConflict, gin.H{
					"status": "A backup destination with this name already exists",
				})
				return
			}

			if isForeignKeyError(err) {
				c.JSON(http.StatusBadRequest, gin.H{
					"status": "Invalid connection ID - connection does not exist",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to create backup destination",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status": "Backup destination created successfully",
			"data":   e,
		})
	}
}

func ListAllBackupDestinations(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var destinations []db.Destination

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		connectionID := c.Query("connection_id")

		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 10
		}

		offset := (page - 1) * limit

		query := conn.Model(&db.Destination{})

		// Filter by connection_id if provided
		if connectionID != "" {
			if connID, err := strconv.ParseUint(connectionID, 10, 32); err == nil {
				query = query.Where("connection_id = ?", uint(connID))
			}
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to count backup destinations",
				"error":  err.Error(),
			})
			return
		}

		if err := query.Offset(offset).Limit(limit).Find(&destinations).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to retrieve backup destinations",
				"error":  err.Error(),
			})
			return
		}

		totalPages := (int(total) + limit - 1) / limit

		c.JSON(http.StatusOK, gin.H{
			"data": destinations,
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
		destinationId := c.Query("destination_id")
		id, err := strconv.ParseUint(destinationId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Invalid ID format",
			})
			return
		}

		var existing db.Destination
		if err := conn.First(&existing, uint(id)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"status": "Backup destination not found",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to retrieve backup destination",
				"error":  err.Error(),
			})
			return
		}

		var updates db.Destination
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Invalid request body",
				"error":  err.Error(),
			})
			return
		}

		if updates.ConnectionID != 0 && updates.ConnectionID != existing.ConnectionID {
			existing.ConnectionID = updates.ConnectionID
		}
		if updates.Name != "" && updates.Name != existing.Name {
			existing.Name = updates.Name
		}
		if updates.EndpointURL != "" {
			existing.EndpointURL = updates.EndpointURL
		}
		if updates.Region != "" {
			existing.Region = updates.Region
		}
		if updates.BucketName != "" {
			existing.BucketName = updates.BucketName
		}
		if updates.AccessKeyID != "" {
			existing.AccessKeyID = updates.AccessKeyID
		}
		if updates.SecretAccessKey != "" {
			existing.SecretAccessKey = updates.SecretAccessKey
		}
		if updates.PathPrefix != existing.PathPrefix {
			existing.PathPrefix = updates.PathPrefix
		}

		existing.UseSSL = updates.UseSSL
		existing.VerifySSL = updates.VerifySSL

		if err := conn.Save(&existing).Error; err != nil {
			if isDuplicateKeyError(err) {
				c.JSON(http.StatusConflict, gin.H{
					"status": "A backup destination with this name already exists",
				})
				return
			}

			if isForeignKeyError(err) {
				c.JSON(http.StatusBadRequest, gin.H{
					"status": "Invalid connection ID - connection does not exist",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to update backup destination",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Backup destination updated successfully",
			"data":    existing,
		})
	}
}

func DeleteBackupDestination(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		destinationId := c.Query("destination_id")
		id, err := strconv.ParseUint(destinationId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Invalid ID format",
			})
			return
		}

		var destination db.Destination
		if err := conn.First(&destination, uint(id)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"status": "Backup destination not found",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to retrieve backup destination",
				"error":  err.Error(),
			})
			return
		}

		if err := conn.Delete(&destination).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to delete backup destination",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
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
