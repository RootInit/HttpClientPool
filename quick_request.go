package HttpClientPool

import (
	"io"
	"net/http"
	"os"
  "HttpPool/Utils"
)

// RequestData represents request data to be passed to QuickRequest
type RequestData struct {
	// Type specifies the HTTP request method (e.g., GET, POST).
	Type string

	// Url is the URL of the HTTP request.
	Url string

	// Params contains the url parameters for the request.
	Params map[string]string

	// JsonData accepts any type with data to be sent in the request body as JSON.
	//
	// Cannot be used with FormData, FormFiles or RawData
	JsonData interface{}

	// FormData contains the data to be sent in the request body as form data.
	//
	// Cannot be used with JsonData or RawData
	FormData map[string]string
	// Files contains the files to be included as part of FormData.
	//
	// Cannot be used with JsonData or RawData
	FormFiles map[string]*os.File

	// RawData contains the raw request body as an io.Reader.
	//
	// Overrides both JsonData and FormData/FormFiles
	RawData *io.Reader

	// Headers contains the HTTP headers for the request.
	Headers map[string]string

	// Cookies contains the cookies to be included in the request.
	Cookies map[string]string
}

// ResponseData represents data from an http.Response returned by QuickRequest
type ResponseData struct {
	// Status is the human-readable status message of the HTTP response.
	Status string

	// StatusCode is the HTTP status code of the response.
	StatusCode int

	// Body contains the raw body of the HTTP response.
	Body []byte

	// Cookies contains the cookies received in the HTTP response.
	Cookies map[string]string
}
// QuickRequest is a convenience wrapper arround http.Request allowing easy basic requests.
// Performs an HTTP request with various options and returns the response.
//
// It allows making HTTP requests with different methods (GET, POST, etc.) and supports request
// options such as URL parameters, headers, user agent, cookies, and various payload types
// including JSON, form, and other data using the RawData field.
//
// Parameters:
//   - request (RequestData): The RequestData struct containing HTTP request data.
//
// Returns:
//   - ResponseData: A ResponseData struct containing HTTP response data.
//   - error: An error, if any, encountered during the HTTP request.
//
func (client *Client) QuickRequest(request RequestData) (ResponseData, error) {
	// Set client status
	client.running = true
	defer func() { client.running = false }()
	// Initialize return variable
	var response = ResponseData{}
	// Set the request body
	var bodyReader io.Reader
	if request.RawData != nil {
		// Use RawData
		bodyReader = *request.RawData
	} else if request.JsonData != nil {
		// Use JsonData
		reader, err := Utils.JsonDataReader(request.JsonData)
		if err != nil {
			return response, err
		}
		bodyReader = reader
	} else if request.FormData != nil || request.FormFiles != nil {
		// Use FormData
		reader, err := Utils.FormDataReader(request.FormData, request.FormFiles)
		if err != nil {
			return response, err
		}
		bodyReader = reader
	} else {
		bodyReader = http.NoBody
	}

	// Create the request
	req, err := http.NewRequest(request.Type, request.Url, bodyReader)
	if err != nil {
		return response, err
	}
	// Set url paramaters
	if request.Params != nil {
		q := req.URL.Query()
		for key, value := range request.Params {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}
	// Set headers
	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}
	// Set userAgent header
	req.Header.Set("user-agent", client.UserAgent)
	// Set cookies
	for name, value := range request.Cookies {
		cookie := http.Cookie{
			Name:  name,
			Value: value,
		}
		req.AddCookie(&cookie)
	}
	// Run request
	res, err := client.Do(req)
	if err != nil {
		return response, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		// Read the body data
		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			return response, err
		}
		resCookies := res.Cookies()
		cookies := make(map[string]string, len(resCookies))
		for _, c := range res.Cookies() {
			cookies[c.Name] = c.Value
		}
		response = ResponseData{
			Status:     res.Status,
			StatusCode: res.StatusCode,
			Body:       responseBody,
			Cookies:    cookies,
		}

		return response, err
	} else if res.StatusCode == http.StatusTooManyRequests {
		// Return RateLimitedError
		return response, &RateLimitedError{}
	} else {
		// Log request information
		return response, &ResponseStatusError{}
	}
}

