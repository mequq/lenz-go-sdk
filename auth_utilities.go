package lenzsdk

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mequq/lenz-go-sdk/entities"
	"github.com/mequq/lenz-go-sdk/logger"
)

// CheckAuthorizationHeaderWithValidUser check request has valid token
func CheckAuthorizationHeaderWithValidUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.Request.Header.Get("Authorization")) == 0 {
			logger.WithRequestHeaders(c).Warn().
				Uint32("logCode", 230101).
				Str("action", "CheckHeaderWithValidUser").
				Msg("Authorization Header is empty")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		// Check and Parse Authorization header token
		claims, err := ParseJWTHeader(c.Request.Header.Get("Authorization"))
		if err != nil {
			logger.WithRequestHeaders(c).Warn().
				Uint32("logCode", 230102).
				Str("action", "CheckHeaderWithValidUser").
				Msg(err.Error())

			c.JSON(http.StatusUnauthorized, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		// Add some headers from JWT token
		addRequiredHeadersFromJWT(c, claims)

		if claims["is_guest"] == true {
			logger.WithRequestHeaders(c).Warn().
				Uint32("logCode", 230103).
				Str("action", "CheckHeaderWithValidUser").
				Msg("The token is for Guest User")

			c.Header("Is-Guest", "True")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "لطفا لاگین کنید"})
			c.Abort()
			return
		}

		// Check if user IP changed then return 401/unauthorize in response
		if checkIfUserIPChanged(c, fmt.Sprintf("%v", claims["ip"])) {
			logger.WithRequestHeaders(c).Warn().
				Uint32("logCode", 230104).
				Str("action", "CheckHeaderWithValidUser").
				Str("previousIP", fmt.Sprintf("%v", claims["ip"])).
				Msg("The user IP has been changed")

			return
		}

		logger.WithRequestHeaders(c).Debug().
			Uint32("logCode", 130100).
			Str("action", "CheckHeaderWithValidUser").
			Msg("The Token is valid")

		c.Next()
	}
}

// CheckProcessableHeaderWithValidUser check request has valid token and be processable
func CheckProcessableHeaderWithValidUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.Request.Header.Get("Authorization")) == 0 {
			logger.WithRequestHeaders(c).Warn().
				Uint32("logCode", 230111).
				Str("action", "CheckHeaderWithValidUserInBackgroundAPI").
				Msg("Authorization Header is empty")

			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		// Check and Parse Authorization header token
		claims, err := ParseJWTHeader(c.Request.Header.Get("Authorization"))
		if err != nil {
			logger.WithRequestHeaders(c).Warn().
				Uint32("logCode", 230112).
				Str("action", "CheckHeaderWithValidUserInBackgroundAPI").
				Msg(err.Error())

			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		// Add some headers from JWT token
		addRequiredHeadersFromJWT(c, claims)

		if claims["is_guest"] == true {
			logger.WithRequestHeaders(c).Warn().
				Uint32("logCode", 230113).
				Str("action", "CheckHeaderWithValidUserInBackgroundAPI").
				Msg("The token is for Guest User")

			c.Header("Is-Guest", "True")
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "لطفا لاگین کنید"})
			c.Abort()
			return
		}

		clientIP := fmt.Sprintf("%v", claims["ip"])
		if clientIP != c.Request.Header.Get(entities.XForwardedForKey) || len(clientIP) == 0 {
			logger.WithRequestHeaders(c).Warn().
				Uint32("logCode", 230114).
				Str("action", "CheckHeaderWithValidUserInBackgroundAPI").
				Str("previousIP", fmt.Sprintf("%v", claims["ip"])).
				Msg("The user IP has been changed")

			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		logger.WithRequestHeaders(c).Debug().
			Uint32("logCode", 130110).
			Str("action", "CheckHeaderWithValidUserInBackgroundAPI").
			Msg("The Token is valid")

		c.Next()
	}
}

// CheckAuthorizationHeaderWithValidOrGuestUser check request has valid token for valid users or loging in  as guest user and returns the token
func CheckAuthorizationHeaderWithValidOrGuestUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Is-Open-Api", "True")

		// Login as guest if Authorization header not included
		if len(c.Request.Header.Get("Authorization")) == 0 {
			_, err := GuestLogin(c)
			if err != nil {
				logger.WithRequestHeaders(c).Error().
					Uint32("logCode", 230121).
					Str("action", "CheckHeaderWithValidOrGuestUser").
					Msg(err.Error())

				c.Abort()
				return
			}

			logger.WithRequestHeaders(c).Debug().
				Uint32("logCode", 130122).
				Str("action", "CheckHeaderWithValidOrGuestUser").
				Msg("Guest Login Successfully")
		}

		// Check and Parse Authorization header token
		claims, err := ParseJWTHeader(c.Request.Header.Get("Authorization"))
		if err != nil {
			logger.WithRequestHeaders(c).Warn().
				Uint32("logCode", 230123).
				Str("action", "CheckHeaderWithValidOrGuestUser").
				Msg(err.Error())

			c.JSON(http.StatusUnauthorized, gin.H{"message": "دسترسی شما منقضی شده است"})
			c.Abort()
			return
		}

		if claims["is_guest"] == true {
			c.Header("Is-Guest", "True")
		}

		// Add some headers from JWT token
		addRequiredHeadersFromJWT(c, claims)

		clientIP := fmt.Sprintf("%v", claims["ip"])
		// if user was a guest we call the guest login again else we return 401
		if clientIP != c.Request.Header.Get(entities.XForwardedForKey) || len(clientIP) == 0 {
			if claims["is_guest"] == false {
				logger.WithRequestHeaders(c).Warn().
					Uint32("logCode", 230124).
					Str("action", "CheckHeaderWithValidOrGuestUser").
					Str("previousIP", clientIP).
					Msg("The user IP has been changed")

				c.JSON(http.StatusUnauthorized, gin.H{"message": "دسترسی شما منقضی شده است"})
				c.Abort()
				return
			}

			_, err := GuestLogin(c)
			if err != nil {
				logger.WithRequestHeaders(c).Error().
					Uint32("logCode", 230125).
					Str("action", "CheckHeaderWithValidOrGuestUser").
					Msg(err.Error())

				c.Abort()
				return
			}

			logger.WithRequestHeaders(c).Debug().
				Uint32("logCode", 130126).
				Str("action", "CheckHeaderWithValidOrGuestUser").
				Msg("Guest Login Successfully")

			claims, err = ParseJWTHeader(c.Request.Header.Get("Authorization"))
			if err != nil {
				logger.WithRequestHeaders(c).Warn().
					Uint32("logCode", 230127).
					Str("action", "CheckHeaderWithValidOrGuestUser").
					Msg(err.Error())

				c.JSON(http.StatusUnauthorized, gin.H{"message": "دسترسی شما منقضی شده است"})
				c.Abort()
				return
			}
		}

		// Add some headers from JWT token
		addRequiredHeadersFromJWT(c, claims)

		logger.WithRequestHeaders(c).Debug().
			Uint32("logCode", 130120).
			Str("action", "CheckHeaderWithValidOrGuestUser").
			Msg("The Token is valid")

		c.Next()
	}
}

func checkIfUserIPChanged(c *gin.Context, clientIP string) bool {
	headerIP := c.Request.Header.Get(entities.XForwardedForKey)

	if clientIP != headerIP || len(clientIP) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "دسترسی شما منقضی شده است"})
		c.Abort()
		return true
	}

	return false
}

// ParseJWTHeader parse jwt header with JWT_SECRET_KEY that comes from .env
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

func addRequiredHeadersFromJWT(c *gin.Context, claims map[string]interface{}) {
	c.Request.Header.Set("MSISDN", fmt.Sprintf("%v", claims["user_id"]))
	c.Request.Header.Set("Token-Id", fmt.Sprintf("%v", claims["token_id"]))
}
