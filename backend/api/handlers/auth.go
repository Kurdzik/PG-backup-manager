package handlers

import (
	"fmt"
	"log"
	"net/http"
	"pg_bckup_mgr/auth"
	"pg_bckup_mgr/db"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateUser(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("CreateUser handler called")

		var r db.User

		err := c.ShouldBindJSON(&r)
		if err != nil {
			log.Printf("Error binding JSON in CreateUser: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Invalid request body",
				"error":  err.Error(),
			})
			return
		}
		log.Printf("CreateUser request: Username=%s", r.Username)

		hashedPassword, err := auth.HashPassword(r.Password)
		if err != nil {
			log.Printf("Error hashing password in CreateUser: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Unexpected error occured",
				"error":  err.Error(),
			})
			return
		}
		log.Printf("Password hashed successfully for user: %s", r.Username)

		r.Password = hashedPassword

		err = db.CreateUser(conn, r)
		if err != nil {
			log.Printf("Error creating user in database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Unexpected error occured",
				"error":  err.Error(),
			})
			return
		}

		log.Printf("User created successfully: %s", r.Username)
		c.JSON(http.StatusOK, gin.H{
			"status": "User created succesfully",
		})
	}
}

func LoginUser(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("LoginUser handler called")

		var r db.User

		err := c.ShouldBindJSON(&r)
		if err != nil {
			log.Printf("Error binding JSON in LoginUser: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Invalid request body",
				"error":  err.Error(),
			})
			return
		}
		user, err := db.GetUserByName(conn, r.Username)

		log.Printf("LoginUser request: ID=%d, Username=%s", user.ID, r.Username)

		if err != nil {
			log.Printf("Error getting user by ID in LoginUser: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Unexpected error occured",
				"error":  err.Error(),
			})
			return
		}
		log.Printf("Retrieved user from database: %s", user.Username)

		err = auth.ValidatePassword(r.Password, user.Password)
		if err != nil {
			log.Printf("Password validation failed for user %s: %v", r.Username, err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": "Wrong Password",
				"error":  err.Error(),
			})
			return
		}
		log.Printf("Password validated successfully for user: %s", r.Username)

		jwtToken, err := auth.CreateJWT(r.Username, 3600*time.Hour)
		if err != nil {
			log.Printf("Error creating JWT token for user %s: %v", r.Username, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Unexpected error occured",
				"error":  err.Error(),
			})
			return
		}
		log.Printf("JWT token created successfully for user: %s", r.Username)

		log.Printf("User %s logged in successfully", r.Username)
		c.JSON(http.StatusOK, gin.H{
			"status":  fmt.Sprintf("User %s logged in", r.Username),
			"payload": jwtToken,
		})
	}
}

func ListUsers(conn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []db.User

		result := conn.Find(&users)

		if result.Error != nil {
			log.Printf("Error during users listing")
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Unexpected error occured",
				"error":  result.Error.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
			"data":   users,
		})
	}
}
