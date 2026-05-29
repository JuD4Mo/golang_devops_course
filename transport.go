package main

import "net/http"

type JWTTransport struct {
	transport http.RoundTripper
	token     string
}

func (j JWTTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if j.token != "" {
		req.Header.Add("Authorization", "Bearer "+j.token)
	}
	return j.transport.RoundTrip(req)
}
