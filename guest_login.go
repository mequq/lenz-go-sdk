package lenzsdk

import (
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func GuestLogin(c *gin.Context) (interface{}, error) {
	deviceType := c.Request.Header.Get("Device-Type")
	if len(deviceType) == 0 {
		deviceType = "WEB"
	}

	clientIP := c.Request.Header.Get("X-Forwarded-For")
	if !IPValidator(clientIP) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Some of the required headers are missed"})
		return nil, errors.New("Some of the required headers are missed")
	}

	url := os.Getenv("GUEST_LOGIN_URL")
	req, err := http.NewRequest("POST", url, nil)
	req.Header.Set("X-Forwarded-For", clientIP)
	req.Header.Set("Device-Type", deviceType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "error while guest login"})
		return nil, err
	}

	authorization := resp.Header.Get("Authorization")
	if len(authorization) > 0 {
		c.Request.Header.Set("Authorization", authorization)
		c.Request.Header.Set("New-Guest-Token", "True")
		return resp, nil
	}

	return nil, errors.New("Try again later")
}
