package middleware

import (
	"fmt"
	"log"
	"net/http"
	"pg_bckup_mgr/auth"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Api-Key ,accept, origin, Cache-Control, X-Requested-With, Token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		jwt := strings.Split(authHeader, " ")[1]

		if jwt == "" {
			log.Println("JWT token not provided")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "JWT token not provided"})
			c.Abort()
			return
		}

		claims, err := auth.ValidateJWT(jwt)
		if err != nil {
			log.Println("Invalid JWT token: ", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid JWT token: %s", err.Error())})
			c.Abort()
			return
		}

		c.Set("FullUserName", claims.Username)
		c.Set("Content-Type", "application/json")

		c.Next()
	}
}
