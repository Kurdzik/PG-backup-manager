package handlers

import (
	"net/http"
	"strconv"

	"pg_bckup_mgr/auth"
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

		decryptedPassword, _ := auth.DecryptPassword(creds.PostgresPassword)
		bckupManager := bck.BackupManager{
			Host:     creds.PostgresHost,
			Port:     creds.PostgresPort,
			DBName:   creds.PostgresDBName,
			User:     creds.PostgresUser,
			Password: decryptedPassword,
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

		decryptedPassword, _ := auth.DecryptPassword(creds.PostgresPassword)
		bckupManager := bck.BackupManager{
			Host:     creds.PostgresHost,
			Port:     creds.PostgresPort,
			DBName:   creds.PostgresDBName,
			User:     creds.PostgresUser,
			Password: decryptedPassword,
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

func DeleteBackup(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		databaseId, err := strconv.Atoi(c.Query("database_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Invalid database_id parameter",
			})
			return
		}

		backupDestination := c.Query("destination")
		if backupDestination == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "destination parameter is required",
			})
			return
		}

		filename := c.Query("filename")
		if filename == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "filename parameter is required",
			})
			return
		}

		creds, err := db.GetCredentialsById(conn, databaseId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}
		decryptedPassword, _ := auth.DecryptPassword(creds.PostgresPassword)
		bckupManager := bck.BackupManager{
			Host:     creds.PostgresHost,
			Port:     creds.PostgresPort,
			DBName:   creds.PostgresDBName,
			User:     creds.PostgresUser,
			Password: decryptedPassword,
		}

		err = bckupManager.DeleteBackup(bck.BackupDestination(backupDestination), filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
			"msg":    "Backup deleted successfully",
		})

	}
}
