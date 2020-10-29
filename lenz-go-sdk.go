package lenzsdk

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// CheckAuthorizationHeaderWithValidUser check request has valid token
func CheckAuthorizationHeaderWithValidUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.Request.Header.Get("Authorization")) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		// Check Authorization header is valid
		authorization := strings.Split(c.Request.Header.Get("Authorization"), "Bearer ")
		if len(authorization) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		// Parse JWT token
		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(authorization[1], claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		// Check if user IP changed then return 401/unauthorize in response
		if checkIfUserIPChanged(c, fmt.Sprintf("%v", claims["ip"])) {
			return
		}

		// Add some headers from JWT token
		c.Request.Header.Set("MSISDN", fmt.Sprintf("%v", claims["user_id"]))

		c.Next()
	}
}

// CheckProcessableHeaderWithValidUser check request has valid token and be processable
func CheckProcessableHeaderWithValidUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.Request.Header.Get("Authorization")) == 0 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		// Check Authorization header is valid
		authorization := strings.Split(c.Request.Header.Get("Authorization"), "Bearer ")
		if len(authorization) != 2 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		// Parse JWT token
		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(authorization[1], claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})

		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		// Check if user IP changed then return 401/unauthorize in response
		if len(c.Request.Header.Get("X-Forwarded-For")) == 0 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Some of the required headers are missed"})
			c.Abort()
			return
		}

		if fmt.Sprintf("%v", claims["ip"]) != c.Request.Header.Get("X-Forwarded-For") {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		// Add some headers from JWT token
		c.Request.Header.Set("MSISDN", fmt.Sprintf("%v", claims["user_id"]))

		c.Next()
	}
}

func checkIfUserIPChanged(c *gin.Context, ClientIP string) bool {
	if len(c.Request.Header.Get("X-Forwarded-For")) == 0 {
		c.JSON(http.StatusForbidden, gin.H{"message": "Some of the required headers are missed"})
		c.Abort()
		return true
	}

	if ClientIP != c.Request.Header.Get("X-Forwarded-For") {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "دسترسی شما منقضی شده است"})
		c.Abort()
		return true
	}

	return false
}
