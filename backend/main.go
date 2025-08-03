package main

import (
	"log"
	"pg_bckup_mgr/api/handlers"
	m "pg_bckup_mgr/api/middleware"
	backup_manager "pg_bckup_mgr/backup-manager"
	"pg_bckup_mgr/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env loaded")
	}

	log.Println("Application Initialization Phase Complete!")
}

func main() {

	r := gin.Default()
	r.Use(m.CORSMiddleware())

	api := r.Group("/api/v1")

	dbConn, err := db.Connect()
	if err != nil {
		log.Panicln("Unable to connect to the database")
	}

	// Register all existing backup schedules at startup
	backup_manager.RegisterBackupSchedules(dbConn)
	log.Println("Backup schedules registered successfully!")

	api.GET("/healthcheck", handlers.Healthcheck())

	// Backup endpoints
	api.POST("/backup/create", handlers.CreateBackup(dbConn))
	api.POST("/backup/restore", handlers.RestoreFromBackup(dbConn))
	api.GET("/backup/list", handlers.ListBackups(dbConn))
	api.DELETE("/backup/delete", handlers.DeleteBackup(dbConn))

	// Backup destination endpoints
	api.POST("/backup-destinations/s3/create", handlers.CreateBackupDestination(dbConn))
	api.GET("/backup-destinations/s3/list", handlers.ListAllBackupDestinations(dbConn))
	api.PUT("/backup-destinations/s3/update", handlers.UpdateBackupDestination(dbConn))
	api.DELETE("/backup-destinations/s3/delete", handlers.DeleteBackupDestination(dbConn))

	// Connection endpoints
	api.POST("/connections/create", handlers.CreateConnection(dbConn))
	api.GET("/connections/list", handlers.ListConnections(dbConn))
	api.PUT("/connections/update", handlers.UpdateConnection(dbConn))
	api.DELETE("/connections/delete", handlers.DeleteConnection(dbConn))

	// Backup schedule endpoints
	api.POST("/schedules/create", handlers.CreateSchedule(dbConn))
	api.GET("/schedules/list", handlers.ListSchedules(dbConn))
	api.GET("/schedules/get", handlers.GetSchedule(dbConn))
	api.PUT("/schedules/update", handlers.UpdateSchedule(dbConn))
	api.DELETE("/schedules/delete", handlers.DeleteSchedule(dbConn))
	api.POST("/schedules/enable", handlers.EnableSchedule(dbConn))
	api.POST("/schedules/disable", handlers.DisableSchedule(dbConn))

	log.Println("🚀 Application Startup Complete! 🚀")
	r.Run(":8080")

}
