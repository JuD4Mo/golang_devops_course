package api

import (
	"net/http"
)

type JWTTransport struct {
	transport  http.RoundTripper
	token      string
	password   string
	loginURL   string
	HTTPClient IClient
}

func (j *JWTTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if j.token == "" {
		if j.password != "" {
			token, err := doLoginRequest(j.HTTPClient, j.loginURL, j.password)
			if err != nil {
				return nil, err
			}
			j.token = token
		}
	}
	if j.token != "" {
		req.Header.Add("Authorization", "Bearer "+j.token)
	}
	return j.transport.RoundTrip(req)
}
