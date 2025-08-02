package handlers

import (
	"net/http"
	"strconv"

	bck "pg_bckup_mgr/backup-manager"
	"pg_bckup_mgr/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateBackupRequest struct {
	DatabaseId  int    `json:"database_id"`
	Destination string `json:"backup_destination"`
}

func CreateBackup(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		var r CreateBackupRequest

		err := c.ShouldBindJSON(&r)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}

		creds, err := db.GetCredentialsById(conn, r.DatabaseId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}

		bckupManager := bck.BackupManager{
			Host:     creds.PostgresHost,
			Port:     creds.PostgresPort,
			DBName:   creds.PostgresDBName,
			User:     creds.PostgresUser,
			Password: creds.PostgresPassword,
		}

		err = bckupManager.CreateBackup(bck.BackupDestination(r.Destination))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})

	}
}

func ListBackups(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		databaseId, _ := strconv.Atoi(c.Query("database_id"))
		backupDestination := c.Query("backup_destination")

		creds, err := db.GetCredentialsById(conn, databaseId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}

		bckupManager := bck.BackupManager{
			Host:     creds.PostgresHost,
			Port:     creds.PostgresPort,
			DBName:   creds.PostgresDBName,
			User:     creds.PostgresUser,
			Password: creds.PostgresPassword,
		}

		files := bckupManager.ListAvaiableBackups(bck.BackupDestination(backupDestination))

		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
			"msg":    files,
		})

	}
}

type RestoreFromBackupRequest struct {
	DatabaseId  int    `json:"database_id"`
	Destination string `json:"backup_destination"`
	Filename    string `json:"backup_filename"`
}

func RestoreFromBackup(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		var r RestoreFromBackupRequest

		err := c.ShouldBindJSON(&r)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}

		creds, err := db.GetCredentialsById(conn, r.DatabaseId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}

		bckupManager := bck.BackupManager{
			Host:     creds.PostgresHost,
			Port:     creds.PostgresPort,
			DBName:   creds.PostgresDBName,
			User:     creds.PostgresUser,
			Password: creds.PostgresPassword,
		}

		err = bckupManager.RestoreFromBackup(bck.BackupDestination(r.Destination), r.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})

	}
}
