package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Response interface {
	GetResponse() string
}

type Page struct {
	Name string `json:"page"`
}

type Words struct {
	Input string   `json:"input"`
	Words []string `json:"words"`
}

func (w Words) GetResponse() string {
	return fmt.Sprintf("%s", strings.Join(w.Words, ", "))
}

type Occurrence struct {
	Words map[string]int `json:"words"`
}

func (o Occurrence) GetResponse() string {
	out := []string{}

	for word, occurrence := range o.Words {
		out = append(out, fmt.Sprintf("%s (%d)", word, occurrence))
	}
	return fmt.Sprintf("%s", strings.Join(out, ", "))
}

func main() {
	var (
		requestURL string
		password   string
		parsedURL  *url.URL
		err        error
	)

	flag.StringVar(&requestURL, "url", "", "url to access")
	flag.StringVar(&password, "password", "", "use a password to access our api")

	flag.Parse()

	if parsedURL, err = url.ParseRequestURI(requestURL); err != nil {
		fmt.Printf("validation error: URL is not valid: %s\n", err)
		flag.Usage()
		os.Exit(1)
	}

	client := http.Client{}

	if password != "" {
		token, err := doLoginRequest(client, parsedURL.Scheme+"://"+parsedURL.Host+"/login", password)
		if err != nil {
			if requestErr, ok := err.(RequestError); ok {
				fmt.Printf("Error: %s (HTTP Code: %d, Body: %s)\n", requestErr.Err, requestErr.HTTPCode, requestErr.Body)
				os.Exit(1)
			}
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		client.Transport = JWTTransport{
			transport: http.DefaultTransport,
			token:     token,
		}
	}

	res, err := doRequest(client, parsedURL.String())
	if err != nil {
		if requestErr, ok := err.(RequestError); ok {
			fmt.Printf("Error: %s (HTTP Code: %d, Body: %s)\n", requestErr.Err, requestErr.HTTPCode, requestErr.Body)
			os.Exit(1)
		}
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	if res == nil {
		fmt.Printf("no response\n")
		os.Exit(1)
	}

	fmt.Printf("response: %s\n", res.GetResponse())
}

func doRequest(client http.Client, requestURL string) (Response, error) {

	response, err := client.Get(requestURL)

	if err != nil {
		return nil, fmt.Errorf("ReadAll error: %s", err)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %s", err)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid output (HTTP Code %d): %s\n", response.StatusCode, string(body))
	}

	if !json.Valid(body) {
		return nil, RequestError{
			HTTPCode: response.StatusCode,
			Body:     string(body),
			Err:      "no valid JSON returned",
		}
	}

	var page Page

	err = json.Unmarshal(body, &page)
	if err != nil {
		return nil, RequestError{
			HTTPCode: response.StatusCode,
			Body:     string(body),
			Err:      fmt.Sprintf("Page Unmarshall error: %s", err),
		}
	}

	switch page.Name {
	case "words":
		var words Words

		err = json.Unmarshal(body, &words)
		if err != nil {
			return nil, RequestError{
				HTTPCode: response.StatusCode,
				Body:     string(body),
				Err:      fmt.Sprintf("Words Unmarshall error: %s", err),
			}
		}
		return words, nil

	case "occurrence":
		var occurence Occurrence

		err = json.Unmarshal(body, &occurence)
		if err != nil {
			return nil, RequestError{
				HTTPCode: response.StatusCode,
				Body:     string(body),
				Err:      fmt.Sprintf("Occurrences Unmarshall error: %s", err),
			}
		}

		return occurence, nil
	}
	return nil, nil
}
