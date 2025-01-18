package global

import (
	"io"
	"net/http"
	"time"
)

var HttpClient = http.Client{
	Timeout: time.Duration(6) * time.Second,
}

func NewHttpRequest(method string, url string, body io.Reader) (*http.Request, error) {
	var req, e = http.NewRequest(method, url, body)
	if e != nil {
		return req, e
	}
	req.Header.Set("User-Agent", "vintheweirdass-assistbot")
	return req, nil
}
