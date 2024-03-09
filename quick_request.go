package HttpClientPool

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// RequestData represents request data to be passed to QuickRequest
type RequestData struct {
	// Type specifies the HTTP request method (e.g., GET, POST).
	Type string

	// Url is the URL of the HTTP request.
	Url string

	// Params contains the url parameters for the request. Key:Array of values
	Params map[string][]string

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

	// Headers contains the HTTP headers for the request. Key:Array of values
	Headers map[string][]string

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
//   - reqData (RequestData): The RequestData struct containing HTTP request data.
//
// Returns:
//   - ResponseData: A ResponseData struct containing HTTP response data.
//   - error: An error, if any, encountered during the HTTP request.
func (client *Client) QuickRequest(reqData RequestData) (ResponseData, error) {
	// Initialize return variable
	var response = ResponseData{}
	// Set the request body
	var bodyReader io.Reader
	if reqData.RawData != nil {
		// Use RawData
		bodyReader = *reqData.RawData
	} else if reqData.JsonData != nil {
		// Use JsonData
		reader, err := jsonDataReader(reqData.JsonData)
		if err != nil {
			return response, err
		}
		bodyReader = reader
	} else if reqData.FormData != nil || reqData.FormFiles != nil {
		// Use FormData
		reader, err := formDataReader(reqData.FormData, reqData.FormFiles)
		if err != nil {
			return response, err
		}
		bodyReader = reader
	} else {
		bodyReader = http.NoBody
	}
	// Create the request
	req, err := http.NewRequest(reqData.Type, reqData.Url, bodyReader)
	if err != nil {
		return response, err
	}
	// Set url paramaters
	if reqData.Params != nil {
		q := req.URL.Query()
		for key, values := range reqData.Params {
			for _, value := range values {
				q.Add(key, value)
			}
		}
		req.URL.RawQuery = q.Encode()
	}
	// Set headers
	for key, values := range reqData.Headers {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}
	// Set userAgent header
	req.Header.Set("user-agent", client.GetUserAgent())
	// Set cookies
	for name, value := range reqData.Cookies {
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
	return response, nil

}

// jsonDataReader converts a Go data structure into a JSON-formatted io.Reader.
//
// Parameters:
//   - data (interface{}): The Go data structure to be converted to JSON.
//
// Returns:
//   - io.Reader: An io.Reader containing the JSON-encoded data.
//   - error: An error, if any, encountered during the JSON encoding process.
func jsonDataReader(data interface{}) (io.Reader, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	bodyData := bytes.NewBuffer(jsonData)
	return bodyData, nil
}

// formDataReader creates a multipart/form-data io.Reader from a map of key-value pairs
// and a map of files.
//
// Parameters:
//   - data (map[string]string): A map of string key-value pairs representing form fields.
//   - files (map[string]*os.File): A map of files to be included in the request body.
//
// Returns:
//   - io.Reader: An io.Reader containing the multipart/form-data request body.
//   - error: An error, if any, encountered during the construction of the request body.
func formDataReader(data map[string]string, files map[string]*os.File) (io.Reader, error) {
	var formData bytes.Buffer
	writer := multipart.NewWriter(&formData)
	// Set Data
	for field, value := range data {
		if err := writer.WriteField(field, value); err != nil {
			return nil, err
		}
	}
	// Set Files
	for field, file := range files {
		fileField, err := writer.CreateFormFile(field, file.Name())
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(fileField, file)
		if err != nil {
			return nil, err
		}
	}
	return &formData, nil
}
