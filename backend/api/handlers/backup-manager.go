package handlers

import (
	"log"
	"net/http"

	"pg_bckup_mgr/auth"
	backup_manager "pg_bckup_mgr/backup-manager"
	"pg_bckup_mgr/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateBackupRequest struct {
	DatabaseId  string `json:"database_id"`
	Destination string `json:"backup_destination"`
}

type RestoreFromBackupRequest struct {
	DatabaseId  string `json:"database_id"`
	Destination string `json:"backup_destination"`
	Filename    string `json:"backup_filename"`
}

func CreateBackup(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("CreateBackup handler called")

		var destination *db.Destination
		var r CreateBackupRequest

		err := c.ShouldBindJSON(&r)
		if err != nil {
			log.Printf("Error binding JSON in CreateBackup: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}
		log.Printf("CreateBackup request: DatabaseId=%s, Destination=%s", r.DatabaseId, r.Destination)

		creds, err := db.GetCredentialsById(conn, r.DatabaseId)
		if err != nil {
			log.Printf("Error getting credentials in CreateBackup: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}
		log.Printf("Retrieved credentials for database: %s", creds.PostgresDBName)

		if r.Destination != "local" {
			log.Printf("Setting up S3 destination for backup")
			dest, err := db.GetBackupDestinationByID(conn, r.Destination)
			if err != nil {
				log.Printf("Error getting backup destination in CreateBackup: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"status": err.Error(),
				})
				return
			}
			destination = &dest
			r.Destination = "s3"
			log.Printf("S3 destination configured: %s", dest.Name)
		}

		decryptedPassword, _ := auth.DecryptPassword(creds.PostgresPassword)
		bckupManager := backup_manager.BackupManager{
			Host:              creds.PostgresHost,
			Port:              creds.PostgresPort,
			DBName:            creds.PostgresDBName,
			User:              creds.PostgresUser,
			Password:          decryptedPassword,
			BackupDestination: destination,
		}
		log.Printf("BackupManager initialized for %s@%s:%s/%s", creds.PostgresUser, creds.PostgresHost, creds.PostgresPort, creds.PostgresDBName)

		err = bckupManager.CreateBackup(backup_manager.BackupDestination(r.Destination))
		if err != nil {
			log.Printf("Error creating backup: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}

		log.Println("Backup created successfully")
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})

	}
}

func ListBackups(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("ListBackups handler called")

		var destination *db.Destination

		databaseId := c.Query("database_id")
		backupDestination := c.Query("backup_destination")
		log.Printf("ListBackups request: DatabaseId=%s, Destination=%s", databaseId, backupDestination)

		creds, err := db.GetCredentialsById(conn, databaseId)
		if err != nil {
			log.Printf("Error getting credentials in ListBackups: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}
		log.Printf("Retrieved credentials for database: %s", creds.PostgresDBName)

		if backupDestination != "local" {
			log.Printf("Setting up S3 destination for listing backups")
			dest, err := db.GetBackupDestinationByID(conn, backupDestination)
			if err != nil {
				log.Printf("Error getting backup destination in ListBackups: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"status": err.Error(),
				})
				return
			}
			destination = &dest
			backupDestination = "s3"
			log.Printf("S3 destination configured: %s", dest.Name)
		}

		bckupManager := backup_manager.BackupManager{
			Host:              creds.PostgresHost,
			Port:              creds.PostgresPort,
			DBName:            creds.PostgresDBName,
			User:              creds.PostgresUser,
			Password:          creds.PostgresPassword,
			BackupDestination: destination,
		}
		log.Printf("BackupManager initialized for listing backups")

		files := bckupManager.ListAvaiableBackups(backup_manager.BackupDestination(backupDestination))
		log.Printf("Found %d backup files", len(files))

		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
			"msg":    files,
		})
	}
}

func RestoreFromBackup(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("RestoreFromBackup handler called")

		var destination *db.Destination
		var r RestoreFromBackupRequest

		err := c.ShouldBindJSON(&r)
		if err != nil {
			log.Printf("Error binding JSON in RestoreFromBackup: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}
		log.Printf("RestoreFromBackup request: DatabaseId=%s, Destination=%s, Filename=%s", r.DatabaseId, r.Destination, r.Filename)

		creds, err := db.GetCredentialsById(conn, r.DatabaseId)
		if err != nil {
			log.Printf("Error getting credentials in RestoreFromBackup: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}
		log.Printf("Retrieved credentials for database: %s", creds.PostgresDBName)

		if r.Destination != "local" {
			log.Printf("Setting up S3 destination for restore")
			dest, err := db.GetBackupDestinationByID(conn, r.Destination)
			if err != nil {
				log.Printf("Error getting backup destination in RestoreFromBackup: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"status": err.Error(),
				})
				return
			}
			destination = &dest
			r.Destination = "s3"
			log.Printf("S3 destination configured: %s", dest.Name)
		}

		decryptedPassword, _ := auth.DecryptPassword(creds.PostgresPassword)
		bckupManager := backup_manager.BackupManager{
			Host:              creds.PostgresHost,
			Port:              creds.PostgresPort,
			DBName:            creds.PostgresDBName,
			User:              creds.PostgresUser,
			Password:          decryptedPassword,
			BackupDestination: destination,
		}
		log.Printf("BackupManager initialized for restore from %s", r.Destination)

		err = bckupManager.RestoreFromBackup(backup_manager.BackupDestination(r.Destination), r.Filename)
		if err != nil {
			log.Printf("Error restoring from backup: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}

		log.Printf("Successfully restored database from backup: %s", r.Filename)
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})

	}
}

func DeleteBackup(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("DeleteBackup handler called")

		var destination *db.Destination

		databaseId := c.Query("database_id")

		backupDestination := c.Query("destination")
		if backupDestination == "" {
			log.Println("Missing destination parameter in DeleteBackup")
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "destination parameter is required",
			})
			return
		}

		filename := c.Query("filename")
		if filename == "" {
			log.Println("Missing filename parameter in DeleteBackup")
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "filename parameter is required",
			})
			return
		}

		log.Printf("DeleteBackup request: DatabaseId=%s, Destination=%s, Filename=%s", databaseId, backupDestination, filename)

		creds, err := db.GetCredentialsById(conn, databaseId)
		if err != nil {
			log.Printf("Error getting credentials in DeleteBackup: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}
		log.Printf("Retrieved credentials for database: %s", creds.PostgresDBName)

		if backupDestination != "local" {
			log.Printf("Setting up S3 destination for delete")
			dest, err := db.GetBackupDestinationByID(conn, backupDestination)
			if err != nil {
				log.Printf("Error getting backup destination in DeleteBackup: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"status": err.Error(),
				})
				return
			}
			destination = &dest
			backupDestination = "s3"
			log.Printf("S3 destination configured: %s", dest.Name)
		}

		decryptedPassword, _ := auth.DecryptPassword(creds.PostgresPassword)
		bckupManager := backup_manager.BackupManager{
			Host:              creds.PostgresHost,
			Port:              creds.PostgresPort,
			DBName:            creds.PostgresDBName,
			User:              creds.PostgresUser,
			Password:          decryptedPassword,
			BackupDestination: destination,
		}
		log.Printf("BackupManager initialized for delete from %s", backupDestination)

		err = bckupManager.DeleteBackup(backup_manager.BackupDestination(backupDestination), filename)
		if err != nil {
			log.Printf("Error deleting backup: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}

		log.Printf("Successfully deleted backup: %s", filename)
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
			"msg":    "Backup deleted successfully",
		})

	}
}
