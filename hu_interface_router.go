package lenzsdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"git.abanppc.com/lenz-public/lenz-go-sdk/entities"
	"git.abanppc.com/lenz-public/lenz-go-sdk/logger"
	"github.com/gin-gonic/gin"
)

// Router is required data for calling HU
type Router struct {
	EndPoint      string
	Data          map[string]interface{}
	Authorization string
	XForwardedFor string
	MSISDN        string
	RequestID     string
}

// NewRouter create HUInterfaceRouter from gin.Context
func NewRouter(c *gin.Context, endPoint string) *Router {
	router := &Router{
		EndPoint:      endPoint,
		Authorization: c.Request.Header.Get("Authorization"),
		XForwardedFor: c.Request.Header.Get(entities.XForwardedForKey),
		MSISDN:        c.Request.Header.Get("MSISDN"),
		RequestID:     c.Request.Header.Get("X-Request-Id"),
		Data:          map[string]interface{}{},
	}

	return router
}

// Execute send data to HU and check default errors
func (r *Router) Execute(c *gin.Context) ([]byte, error) {
	return r.do(c, false, true, http.StatusNotAcceptable, "در انجام درخواست شما خطایی رخ داده است")
}

// ExecuteWithCriticalErrors send data to HU and fill gin.c when error is critical
func (r *Router) ExecuteWithCriticalErrors(c *gin.Context) ([]byte, error) {
	return r.do(c, false, false, http.StatusNotAcceptable, "در انجام درخواست شما خطایی رخ داده است")
}

// ExecuteBackgroundUseCase send data to HU and check default errors
func (r *Router) ExecuteBackgroundUseCase(c *gin.Context) ([]byte, error) {
	return r.do(c, true, true, http.StatusUnprocessableEntity, "در انجام درخواست شما خطایی رخ داده است")
}

func (r *Router) do(c *gin.Context, backgroundUseCase bool, checkAllErrors bool, errorStatusCode int, errorMessage string) ([]byte, error) {
	payload, err := json.Marshal(r.Data)
	if err != nil {
		c.JSON(errorStatusCode, gin.H{"message": errorMessage})
		return nil, err
	}

	url := os.Getenv("HU_INTERFACE_URL") + r.EndPoint
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", r.Authorization)
	req.Header.Set("X-Forwarded-For", r.XForwardedFor)
	req.Header.Set("MSISDN", r.MSISDN)
	req.Header.Set("X-Request-Id", r.RequestID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.WithRequestHeaders(c).Error().
			Uint32("logCode", 230201).
			Str("action", "HUInterfaceRouter").
			Msg(err.Error())

		if checkAllErrors {
			c.JSON(errorStatusCode, gin.H{"message": errorMessage})
		}

		return nil, err
	}
	defer resp.Body.Close()

	byteResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.WithRequestHeaders(c).Error().
			Uint32("logCode", 230202).
			Str("action", "HUInterfaceRouter").
			Msg(err.Error())

		if checkAllErrors {
			c.JSON(errorStatusCode, gin.H{"message": errorMessage})
		}

		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		if backgroundUseCase {
			if checkAllErrors {
				c.Data(errorStatusCode, gin.MIMEJSON, byteResponse)
			}
		} else {
			c.Data(resp.StatusCode, gin.MIMEJSON, byteResponse)
		}
		return nil, errors.New("client should login again")
	} else if resp.StatusCode != http.StatusOK {
		if checkAllErrors {
			c.JSON(errorStatusCode, gin.H{"message": errorMessage})
		}

		return nil, errors.New("invalid response")
	}

	return byteResponse, nil
}
