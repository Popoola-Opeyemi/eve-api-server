package echotools

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// Get performs an httpGet while retaining the original request header
// when body = true the returned response.Body isnt closed
func Get(c echo.Context, url string, body bool) (resp *http.Response, err error) {
	client := &http.Client{Timeout: time.Second * 10}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// copy current request headers
	for k, v := range c.Request().Header {
		for _, i := range v {
			req.Header.Add(k, i)
		}
	}

	// make request
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	if body == false {
		defer resp.Body.Close()
	}

	return
}

// Post performs an httpGet while retaining the original request header
// when body = true the returned response.Body isnt closed
func Post(c echo.Context, url string, payload interface{}, body bool) (resp *http.Response, err error) {
	client := &http.Client{Timeout: time.Second * 10}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return
	}

	// copy current request headers
	for k, v := range c.Request().Header {
		for _, i := range v {
			req.Header.Add(k, i)
		}
	}

	// set content type
	req.Header.Set("Content-Type", "application/json")

	// make request
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	if body == false {
		defer resp.Body.Close()
	}

	return
}
