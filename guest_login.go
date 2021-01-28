package lenzsdk

import (
	"errors"
	"net/http"
	"os"

	"git.abanppc.com/lenz-public/lenz-go-sdk/entities"
	"github.com/gin-gonic/gin"
)

// GuestLogin login user as guest
func GuestLogin(c *gin.Context) (interface{}, error) {
	deviceType := c.Request.Header.Get("Device-Type")
	if len(deviceType) == 0 {
		deviceType = "WEB"
	}

	clientIP := c.Request.Header.Get(entities.XForwardedForKey)
	if !IPValidator(clientIP) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Some of the required headers are missed"})
		return nil, errors.New("Some of the required headers are missed")
	}

	url := os.Getenv("GUEST_LOGIN_URL")
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "error while create request"})
		return nil, err
	}
	req.Header.Set(entities.XForwardedForKey, clientIP)
	req.Header.Set("Device-Type", deviceType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "error while guest login"})
		return nil, err
	}

	authorization := resp.Header.Get("Authorization")
	if len(authorization) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Authorization header is empty after guest login"})
		return nil, errors.New("Try again later")
	}

	c.Request.Header.Set("Authorization", authorization)
	c.Header("Authorization", authorization)
	c.Header("Is-Guest", "True")
	return resp, nil
}
