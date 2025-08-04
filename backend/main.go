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

	// User auth
	api.POST("/user/create", handlers.CreateUser(dbConn))
	api.POST("/user/login", handlers.LoginUser(dbConn))
	api.GET("/user/list", handlers.ListUsers(dbConn))

	apiProtected := api.Use(m.AuthMiddleware())

	// Backup endpoints
	apiProtected.POST("/backup/create", handlers.CreateBackup(dbConn))
	apiProtected.POST("/backup/restore", handlers.RestoreFromBackup(dbConn))
	apiProtected.GET("/backup/list", handlers.ListBackups(dbConn))
	apiProtected.DELETE("/backup/delete", handlers.DeleteBackup(dbConn))

	// Backup destination endpoints
	apiProtected.POST("/backup-destinations/s3/create", handlers.CreateBackupDestination(dbConn))
	apiProtected.GET("/backup-destinations/s3/list", handlers.ListAllBackupDestinations(dbConn))
	apiProtected.PUT("/backup-destinations/s3/update", handlers.UpdateBackupDestination(dbConn))
	apiProtected.DELETE("/backup-destinations/s3/delete", handlers.DeleteBackupDestination(dbConn))

	// Connection endpoints
	apiProtected.POST("/connections/create", handlers.CreateConnection(dbConn))
	apiProtected.GET("/connections/list", handlers.ListConnections(dbConn))
	apiProtected.PUT("/connections/update", handlers.UpdateConnection(dbConn))
	apiProtected.DELETE("/connections/delete", handlers.DeleteConnection(dbConn))

	// Backup schedule endpoints
	apiProtected.POST("/schedules/create", handlers.CreateSchedule(dbConn))
	apiProtected.GET("/schedules/list", handlers.ListSchedules(dbConn))
	apiProtected.GET("/schedules/get", handlers.GetSchedule(dbConn))
	apiProtected.PUT("/schedules/update", handlers.UpdateSchedule(dbConn))
	apiProtected.DELETE("/schedules/delete", handlers.DeleteSchedule(dbConn))
	apiProtected.POST("/schedules/enable", handlers.EnableSchedule(dbConn))
	apiProtected.POST("/schedules/disable", handlers.DisableSchedule(dbConn))

	log.Println("🚀 Application Startup Complete! 🚀")
	r.Run(":8080")

}
