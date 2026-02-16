package lenzsdk

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mequq/lenz-go-sdk/entities"
	"github.com/mequq/lenz-go-sdk/logger"
)

// TiamoRouter is required data for calling Tiamo
type TiamoRouter struct {
	EndPoint      string
	Method        string
	Data          map[string]interface{}
	XForwardedFor string
	MSISDN        string
	RequestID     string
}

// NewTiamoRouter create TiamoRouter from gin.Context
func NewTiamoRouter(c *gin.Context, method, endPoint string) *TiamoRouter {
	router := &TiamoRouter{
		EndPoint:      endPoint,
		Method:        method,
		XForwardedFor: c.Request.Header.Get(entities.XForwardedForKey),
		MSISDN:        c.Request.Header.Get("MSISDN"),
		RequestID:     c.Request.Header.Get("X-Request-Id"),
		Data:          map[string]interface{}{},
	}

	return router
}

// Execute send data to TiamoWrapper and check default errors
func (r *TiamoRouter) Execute(c *gin.Context) (int, []byte, error) {
	payload, err := json.Marshal(r.Data)
	if err != nil {
		return 0, nil, err
	}

	url := os.Getenv("TW_BASE_URL") + r.EndPoint
	req, err := http.NewRequest(r.Method, url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", r.XForwardedFor)
	req.Header.Set("MSISDN", r.MSISDN)
	req.Header.Set("X-Request-Id", r.RequestID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.WithRequestHeaders(c).Error().
			Uint32("logCode", 230211).
			Str("action", "TiamoWrapperRouter").
			Msg(err.Error())

		return 0, nil, err
	}
	defer resp.Body.Close()

	byteResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.WithRequestHeaders(c).Error().
			Uint32("logCode", 230212).
			Str("action", "TiamoWrapperRouter").
			Msg(err.Error())

		return 0, nil, err
	}

	return resp.StatusCode, byteResponse, nil
}
