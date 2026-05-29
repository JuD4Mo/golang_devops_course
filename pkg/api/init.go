package api

import (
	"io"
	"net/http"
)

type IClient interface {
	Get(url string) (resp *http.Response, err error)
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
}

type Options struct {
	Password string
	LoginURL string
}

type IApi interface {
	DoGetRequest(requestURL string) (Response, error)
}

type Api struct {
	Options Options
	Client  IClient
}

func New(options Options) IApi {
	return Api{
		Options: options,
		Client: &http.Client{
			Transport: &JWTTransport{
				transport:  http.DefaultTransport,
				password:   options.Password,
				loginURL:   options.LoginURL,
				HTTPClient: &http.Client{},
			},
		},
	}
}
