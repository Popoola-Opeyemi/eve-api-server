package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// VerifyPayment ...
func VerifyPayment(uRL, secretKey string) (byteRes []byte, err error) {
	log := Env.Log

	req, err := http.NewRequest(http.MethodGet, uRL, nil)
	if err != nil {
		return
	}

	// setting the header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secretKey))

	// Initializing http client model
	client := &http.Client{}

	// send request
	resp, err := client.Do(req)

	if err != nil {
		return
	}

	// read response as byte
	byteRes, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return
	}

	log.Debug("end of verification")

	return
}
