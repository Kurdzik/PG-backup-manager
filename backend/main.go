package main

import (
	"log"
	"pg_bckup_mgr/api/handlers"
	m "pg_bckup_mgr/api/middleware"
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

	api.GET("/healthcheck", handlers.Healthcheck())

	api.POST("/backup/create", handlers.CreateBackup(dbConn))
	api.POST("/backup/restore", handlers.RestoreFromBackup(dbConn))
	api.GET("/backup/list", handlers.ListBackups(dbConn))

	api.POST("/connections/create", handlers.CreateConnection(dbConn))
	api.GET("/connections/list", handlers.ListConnections(dbConn))
	api.PUT("/connections/update", handlers.UpdateConnection(dbConn))
	api.DELETE("/connections/delete", handlers.DeleteConnection(dbConn))

	log.Println("Application Startup Complete!")
	r.Run(":8080")

}
