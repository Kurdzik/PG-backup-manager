package handlers

import (
	"log"
	"net/http"
	backup_manager "pg_bckup_mgr/backup-manager"
	"pg_bckup_mgr/db"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateScheduleRequest struct {
	ConnectionID  string `json:"connection_id" binding:"required"`
	DestinationID string `json:"destination_id" binding:"required"`
	Schedule      string `json:"schedule" binding:"required"`
}
type UpdateScheduleRequest struct {
	Schedule      *string `json:"schedule,omitempty"`
	Enabled       *bool   `json:"enabled,omitempty"`
	ConnectionID  *string `json:"connection_id,omitempty"`
	DestinationID *string `json:"destination_id,omitempty"`
}

func CreateSchedule(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("CreateSchedule handler called")
		var r CreateScheduleRequest
		err := c.ShouldBindJSON(&r)
		if err != nil {
			log.Printf("Error binding JSON in CreateSchedule: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request format",
				"error":   err.Error(),
			})
			return
		}
		log.Printf("CreateSchedule request: ConnectionID=%s, DestinationID=%s, Schedule=%s",
			r.ConnectionID, r.DestinationID, r.Schedule)
		_, err = db.GetCredentialsById(conn, r.ConnectionID)
		if err != nil {
			log.Printf("Error getting connection in CreateSchedule: %v", err)
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Connection not found",
				"error":   err.Error(),
			})
			return
		}
		_, err = db.GetBackupDestinationByID(conn, r.DestinationID)
		if err != nil {
			log.Printf("Error getting destination in CreateSchedule: %v", err)
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Destination not found",
				"error":   err.Error(),
			})
			return
		}
		err = backup_manager.CreateSchedule(conn, r.ConnectionID, r.DestinationID, r.Schedule)
		if err != nil {
			log.Printf("Error creating schedule: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to create schedule",
				"error":   err.Error(),
			})
			return
		}
		backup_manager.RestartBackupScheduler(conn)
		log.Println("Schedule created successfully")
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Schedule created successfully",
		})
	}
}

func UpdateSchedule(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("UpdateSchedule handler called")
		scheduleID := c.Query("schedule_id")
		if scheduleID == "" {
			log.Println("Missing schedule ID parameter in UpdateSchedule")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Schedule ID is required",
			})
			return
		}
		var r UpdateScheduleRequest
		err := c.ShouldBindJSON(&r)
		if err != nil {
			log.Printf("Error binding JSON in UpdateSchedule: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request format",
				"error":   err.Error(),
			})
			return
		}
		log.Printf("UpdateSchedule request: ScheduleID=%s", scheduleID)
		updates := make(map[string]interface{})
		if r.Schedule != nil {
			updates["schedule"] = *r.Schedule
			log.Printf("Updating schedule to: %s", *r.Schedule)
		}
		if r.Enabled != nil {
			updates["enabled"] = *r.Enabled
			log.Printf("Updating enabled to: %v", *r.Enabled)
		}
		if r.ConnectionID != nil {
			_, err := db.GetCredentialsById(conn, *r.ConnectionID)
			if err != nil {
				log.Printf("Error validating connection in UpdateSchedule: %v", err)
				c.JSON(http.StatusNotFound, gin.H{
					"status":  http.StatusNotFound,
					"message": "Connection not found",
					"error":   err.Error(),
				})
				return
			}
			connID, _ := strconv.ParseUint(*r.ConnectionID, 10, 32)
			updates["connection_id"] = uint(connID)
			log.Printf("Updating connection_id to: %s", *r.ConnectionID)
		}
		if r.DestinationID != nil {
			_, err := db.GetBackupDestinationByID(conn, *r.DestinationID)
			if err != nil {
				log.Printf("Error validating destination in UpdateSchedule: %v", err)
				c.JSON(http.StatusNotFound, gin.H{
					"status":  http.StatusNotFound,
					"message": "Destination not found",
					"error":   err.Error(),
				})
				return
			}
			destID, _ := strconv.ParseUint(*r.DestinationID, 10, 32)
			updates["destination_id"] = uint(destID)
			log.Printf("Updating destination_id to: %s", *r.DestinationID)
		}
		if len(updates) == 0 {
			log.Println("No valid fields to update in UpdateSchedule")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "No valid fields provided for update",
			})
			return
		}
		err = backup_manager.UpdateSchedule(conn, scheduleID, updates)
		if err != nil {
			log.Printf("Error updating schedule: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to update schedule",
				"error":   err.Error(),
			})
			return
		}
		backup_manager.RestartBackupScheduler(conn)
		log.Printf("Schedule %s updated successfully", scheduleID)
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Schedule updated successfully",
		})
	}
}

func DeleteSchedule(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("DeleteSchedule handler called")
		scheduleID := c.Query("schedule_id")
		if scheduleID == "" {
			log.Println("Missing schedule ID parameter in DeleteSchedule")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Schedule ID is required",
			})
			return
		}
		log.Printf("DeleteSchedule request: ScheduleID=%s", scheduleID)
		err := backup_manager.DeleteSchedule(conn, scheduleID)
		if err != nil {
			log.Printf("Error deleting schedule: %v", err)
			if err.Error() == "schedule not found" {
				c.JSON(http.StatusNotFound, gin.H{
					"status":  http.StatusNotFound,
					"message": "Schedule not found",
					"error":   err.Error(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to delete schedule",
				"error":   err.Error(),
			})
			return
		}
		backup_manager.RestartBackupScheduler(conn)
		log.Printf("Schedule %s deleted successfully", scheduleID)
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Schedule deleted successfully",
		})
	}
}

func ListSchedules(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("ListSchedules handler called")
		connectionID := c.Query("connection_id")
		destinationID := c.Query("destination_id")
		enabledStr := c.Query("enabled")
		log.Printf("ListSchedules request: ConnectionID=%s, DestinationID=%s, Enabled=%s",
			connectionID, destinationID, enabledStr)
		filters := make(map[string]interface{})
		if connectionID != "" {
			connID, err := strconv.ParseUint(connectionID, 10, 32)
			if err != nil {
				log.Printf("Invalid connection_id parameter: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  http.StatusBadRequest,
					"message": "Invalid connection_id parameter",
					"error":   err.Error(),
				})
				return
			}
			filters["connection_id"] = uint(connID)
		}
		if destinationID != "" {
			destID, err := strconv.ParseUint(destinationID, 10, 32)
			if err != nil {
				log.Printf("Invalid destination_id parameter: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  http.StatusBadRequest,
					"message": "Invalid destination_id parameter",
					"error":   err.Error(),
				})
				return
			}
			filters["destination_id"] = uint(destID)
		}
		if enabledStr != "" {
			enabled, err := strconv.ParseBool(enabledStr)
			if err != nil {
				log.Printf("Invalid enabled parameter: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  http.StatusBadRequest,
					"message": "Invalid enabled parameter (must be true or false)",
					"error":   err.Error(),
				})
				return
			}
			filters["enabled"] = enabled
		}
		schedules, err := backup_manager.ListSchedules(conn, filters)
		if err != nil {
			log.Printf("Error listing schedules: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to list schedules",
				"error":   err.Error(),
			})
			return
		}
		log.Printf("Found %d schedules", len(schedules))
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "OK",
			"data":    schedules,
			"count":   len(schedules),
		})
	}
}

func GetSchedule(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("GetSchedule handler called")
		scheduleID := c.Query("schedule_id")
		if scheduleID == "" {
			log.Println("Missing schedule ID parameter in GetSchedule")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Schedule ID is required",
			})
			return
		}
		log.Printf("GetSchedule request: ScheduleID=%s", scheduleID)
		schedule, err := backup_manager.GetScheduleByID(conn, scheduleID)
		if err != nil {
			log.Printf("Error getting schedule: %v", err)
			if err.Error() == "schedule not found" {
				c.JSON(http.StatusNotFound, gin.H{
					"status":  http.StatusNotFound,
					"message": "Schedule not found",
					"error":   err.Error(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to get schedule",
				"error":   err.Error(),
			})
			return
		}
		log.Printf("Retrieved schedule %s successfully", scheduleID)
		c.JSON(http.StatusOK, gin.H{
			"status":   http.StatusOK,
			"message":  "OK",
			"schedule": schedule,
		})
	}
}

func EnableSchedule(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("EnableSchedule handler called")
		scheduleID := c.Query("schedule_id")
		if scheduleID == "" {
			log.Println("Missing schedule ID parameter in EnableSchedule")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Schedule ID is required",
			})
			return
		}
		log.Printf("EnableSchedule request: ScheduleID=%s", scheduleID)
		err := backup_manager.EnableSchedule(conn, scheduleID)
		if err != nil {
			log.Printf("Error enabling schedule: %v", err)
			if err.Error() == "schedule not found" {
				c.JSON(http.StatusNotFound, gin.H{
					"status":  http.StatusNotFound,
					"message": "Schedule not found",
					"error":   err.Error(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to enable schedule",
				"error":   err.Error(),
			})
			return
		}
		backup_manager.RestartBackupScheduler(conn)
		log.Printf("Schedule %s enabled successfully", scheduleID)
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Schedule enabled successfully",
		})
	}
}

func DisableSchedule(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("DisableSchedule handler called")
		scheduleID := c.Query("schedule_id")
		if scheduleID == "" {
			log.Println("Missing schedule ID parameter in DisableSchedule")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Schedule ID is required",
			})
			return
		}
		log.Printf("DisableSchedule request: ScheduleID=%s", scheduleID)
		err := backup_manager.DisableSchedule(conn, scheduleID)
		if err != nil {
			log.Printf("Error disabling schedule: %v", err)
			if err.Error() == "schedule not found" {
				c.JSON(http.StatusNotFound, gin.H{
					"status":  http.StatusNotFound,
					"message": "Schedule not found",
					"error":   err.Error(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to disable schedule",
				"error":   err.Error(),
			})
			return
		}
		backup_manager.RestartBackupScheduler(conn)
		log.Printf("Schedule %s disabled successfully", scheduleID)
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Schedule disabled successfully",
		})
	}
}
