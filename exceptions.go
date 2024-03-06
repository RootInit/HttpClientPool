package HttpClientPool

import (
	"fmt"
	"net/http"
)

type RateLimitedError struct{}

func (e *RateLimitedError) Error() string {
	return "429: Too many requests."
}

type ResponseStatusError struct {
	Response http.Response
	Request  RequestData
}

func (e *ResponseStatusError) Error() string {
	message := fmt.Sprintf("%d: %s\n", e.Response.StatusCode, e.Response.Status)
	message += fmt.Sprintf(" %s URL: %v\n", e.Request.Type, e.Request.Params)
	message += fmt.Sprintf(" Body: %v\n", e.Response.Body)
	message += " Params:\n"
	for key, value := range e.Request.Params {
		message += fmt.Sprintf("   %s: %s\n", key, value)
	}
	message += " Headers:\n"
	for key, value := range e.Request.Headers {
		message += fmt.Sprintf("   %s: %s\n", key, value)
	}
	return message
}

type ResponseKeyError struct {
	Map string
	Key interface{}
}

func (e *ResponseKeyError) Error() string {
	return fmt.Sprintf("Key \"%v\" does not exist in %s", e.Key, e.Map)
}
