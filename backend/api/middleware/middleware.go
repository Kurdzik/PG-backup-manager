package middleware

import (
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

// func AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var tokenHeader string

// 		// Check for token in header first
// 		tokenHeader = c.GetHeader("Token")

// 		// If not in header, check query parameter (for WebSocket connections)

// 		if tokenHeader == "" {
// 			tokenHeader = c.Query("token")
// 		}

// 		if os.Getenv("APP_ENV") != "DEV" {

// 			if tokenHeader == "" {
// 				log.Printf("JWT token not provided\n")
// 				c.JSON(401, gin.H{"error": "JWT token not provided"})
// 				c.Abort()
// 				return
// 			}

// 			// Validate and parse token
// 			claims, err := auth.ParseJWT(tokenHeader)
// 			if err != nil {
// 				log.Printf("Invalid or expired JWT token: %v\n", err)
// 				c.JSON(401, gin.H{"error": "Invalid or expired JWT token"})
// 				c.Abort()
// 				return
// 			}

// 			// Check if JWT Token is still valid
// 			if claims.Exp < time.Now().Unix() {
// 				log.Printf("Expired JWT token. Expired at: %v\n", claims.ExpiresAt)
// 				c.JSON(401, gin.H{"error": "Expired JWT token"})
// 				c.Abort()
// 				return
// 			}

// 			c.Set("AuthorizedApps", strings.Join(claims.AllowedPaths, ","))
// 			c.Set("IsValid", true)
// 			c.Set("FullUserName", claims.UserFullName)
// 			c.Set("UserEmail", claims.Username)
// 			c.Set("ADGroups", strings.Join(claims.ADGroups, ","))

// 		}

// 		c.Set("Content-Type", "application/json")

// 		c.Next()
// 	}
// }
