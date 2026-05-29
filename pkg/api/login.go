package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type LoginRequest struct {
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func doLoginRequest(client IClient, requestURL, password string) (string, error) {
	logReq := LoginRequest{
		Password: password,
	}

	body, err := json.Marshal(logReq)
	if err != nil {
		return "", fmt.Errorf("Marshal error: %s", err)
	}

	// We use NewBuffer to make the body ([]byte type) implements the Read() function
	response, err := client.Post(requestURL, "application/json", bytes.NewBuffer(body))

	if err != nil {
		return "nil", fmt.Errorf("http post error: %s", err)
	}

	defer response.Body.Close()

	resBody, err := io.ReadAll(response.Body)

	if err != nil {
		return "", fmt.Errorf("unmarshal error: %s", err)
	}

	if response.StatusCode != 200 {
		return "", fmt.Errorf("Invalid output (HTTP Code %d): %s\n", response.StatusCode, string(resBody))
	}

	if !json.Valid(resBody) {
		return "", RequestError{
			HTTPCode: response.StatusCode,
			Body:     string(resBody),
			Err:      "no valid JSON returned",
		}
	}

	var loginResponse LoginResponse

	err = json.Unmarshal(resBody, &loginResponse)
	if err != nil {
		return "", RequestError{
			HTTPCode: response.StatusCode,
			Body:     string(resBody),
			Err:      fmt.Sprintf("Page Unmarshall error: %s", err),
		}
	}

	if loginResponse.Token == "" {
		return "", RequestError{
			HTTPCode: response.StatusCode,
			Body:     string(resBody),
			Err:      "empty token replied",
		}
	}

	return loginResponse.Token, nil
}
