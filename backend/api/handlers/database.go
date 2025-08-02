package handlers

import (
	"net/http"
	"pg_bckup_mgr/auth"
	"pg_bckup_mgr/db"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateConnectionRequest struct {
	PostgresHost     string `json:"postgres_host" binding:"required"`
	PostgresPort     string `json:"postgres_port" binding:"required"`
	PostgresDBName   string `json:"postgres_db_name" binding:"required"`
	PostgresUser     string `json:"postgres_user" binding:"required"`
	PostgresPassword string `json:"postgres_password" binding:"required"`
}

type UpdateConnectionRequest struct {
	PostgresHost     string `json:"postgres_host"`
	PostgresPort     string `json:"postgres_port"`
	PostgresDBName   string `json:"postgres_db_name"`
	PostgresUser     string `json:"postgres_user"`
	PostgresPassword string `json:"postgres_password"`
}

func CreateConnection(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateConnectionRequest

		testFlag := c.Query("test_connection")

		err := c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "validation error",
				"error":  err.Error(),
			})
			return
		}

		encryptedPassword, err := auth.EncryptPassword(req.PostgresPassword)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "validation error",
				"error":  err.Error(),
			})
			return
		}
		connection := db.Connections{
			PostgresHost:     req.PostgresHost,
			PostgresPort:     req.PostgresPort,
			PostgresDBName:   req.PostgresDBName,
			PostgresUser:     req.PostgresUser,
			PostgresPassword: encryptedPassword,
		}

		if testFlag == "true" {
			if !db.TestConnection(connection) {
				c.JSON(http.StatusRequestTimeout, gin.H{
					"status": "failed to create connection",
					"error":  "Could not reach specified Database",
				})
				return
			} else {
				c.JSON(http.StatusOK, gin.H{
					"status": "Conection Test successful",
				})
				return
			}
		}

		err = db.AddCredentials(conn, connection)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "failed to create connection",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status": "connection created successfully",
			"data":   connection,
		})
	}
}

func ListConnections(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		connections, err := db.ListAllCredentials(conn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "failed to retrieve connections",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
			"data":   connections,
			"count":  len(connections),
		})
	}
}

func UpdateConnection(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		connectionId := c.Query("connection_id")
		id, err := strconv.Atoi(connectionId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "invalid connection ID",
				"error":  err.Error(),
			})
			return
		}

		var req UpdateConnectionRequest
		err = c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "validation error",
				"error":  err.Error(),
			})
			return
		}

		connection, err := db.GetCredentialsById(conn, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "connection not found",
				"error":  err.Error(),
			})
			return
		}

		if req.PostgresHost != "" {
			connection.PostgresHost = req.PostgresHost
		}
		if req.PostgresPort != "" {
			connection.PostgresPort = req.PostgresPort
		}
		if req.PostgresDBName != "" {
			connection.PostgresDBName = req.PostgresDBName
		}
		if req.PostgresUser != "" {
			connection.PostgresUser = req.PostgresUser
		}

		if req.PostgresPassword != "" {
			encryptedPassword, err := auth.EncryptPassword(req.PostgresPassword)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status": "validation error",
					"error":  err.Error(),
				})
				return
			}
			connection.PostgresPassword = encryptedPassword
		}

		err = db.UpdateCredentials(conn, connection)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "failed to update connection",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "connection updated successfully",
			"data":   connection,
		})
	}
}

func DeleteConnection(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		connectionId := c.Query("connection_id")
		id, err := strconv.Atoi(connectionId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "invalid connection ID",
				"error":  err.Error(),
			})
			return
		}

		err = db.DeleteCredentialsById(conn, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "failed to delete connection",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "connection deleted successfully",
		})
	}
}
