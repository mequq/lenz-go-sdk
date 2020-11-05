package lenzsdk

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
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
		if !IPValidator(c.Request.Header.Get("X-Forwarded-For")) {
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

// CheckProcessableHeaderWithValidUser check request has valid token and be processable
func CheckAuthorizationHeaderWithValidOrGuestUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Check X-Forwarded-For IP unvalid then return 401/unauthorize in response
		if !IPValidator(c.Request.Header.Get("X-Forwarded-For")) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Some of the required headers are missed"})
			c.Abort()
			return
		}

		// Login as guest if Authorization header not included
		if len(c.Request.Header.Get("Authorization")) == 0 {
			_, err := GuestLogin(c)
			if err != nil {
				c.Abort()
				return
			}
		}

		// Check Authorization header
		claims, err := ParseJWTHeader(c.Request.Header.Get("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		// if user was a guest we call the guest login again else we return 401
		if fmt.Sprintf("%v", claims["ip"]) != c.Request.Header.Get("X-Forwarded-For") {
			if claims["is_guest"] == false {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "دسترسی شما منقضی شده است"})
				c.Abort()
				return
			} else {
				_, err := GuestLogin(c)
				if err != nil {
					c.Abort()
					return
				}
			}
		}

		// Add some headers from JWT token
		c.Request.Header.Set("MSISDN", fmt.Sprintf("%v", claims["user_id"]))

		c.Next()
	}
}

// IPValidator checks the input string with regex based on real IP Adress Version 4 like 192.168.0.0
func IPValidator(ipAddress string) bool {
	match, _ := regexp.MatchString(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`, ipAddress)
	return match
}

func checkIfUserIPChanged(c *gin.Context, clientIP string) bool {

	headerIP := c.Request.Header.Get("X-Forwarded-For")

	if !IPValidator(headerIP) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Some of the required headers are missed or invalid"})
		c.Abort()
		return true
	}

	if clientIP != headerIP {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "دسترسی شما منقضی شده است"})
		c.Abort()
		return true
	}

	return false
}
func ParseJWTHeader(token string) (map[string]interface{}, error) {
	authorization := strings.Split(token, "Bearer ")
	if len(authorization) != 2 {
		return nil, errors.New("Not valid token")
	}

	// Parse JWT token
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(authorization[1], claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})

	if err != nil {
		return nil, err
	}

	return claims, nil
}
