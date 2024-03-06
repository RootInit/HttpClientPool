package HttpClientPool

import (
	"HttpPool/Utils"
	"encoding/json"
	"os"
	"testing"
)

type PostmanEchoResponse struct {
	URL     string                 `json:"url"`
	Args    map[string]string      `json:"args"`
	Headers map[string]string      `json:"headers"`
	Cookies map[string]string      `json:"cookies"`
	Method  string                 `json:"method"`
	JSON    interface{}            `json:"json"`
	Files   map[string]interface{} `json:"files"`
}

func TestAPICall(t *testing.T) {
	const clientDelayMs = 500
	const maxRPS = 25
	clientDelay := Utils.MillisecondToDuration(clientDelayMs)
	poolDelay := Utils.RpsToDuration(maxRPS)
	// Create pool
	pool := NewClientPool(clientDelay, poolDelay, nil, map[string]float32{"HttpPoolClient": 1})
	client := pool.GetClient()
	// Get Request Test
	request := RequestData{
		Type: "GET",
		Url:  "https://postman-echo.com/get",
		Params: map[string]string{
			"one": "1",
			"two": "2",
		},
		Headers: map[string]string{
			"headkey": "1",
		},
	}
	responseData, err := client.QuickRequest(request)
	if err != nil {
		t.Fatal(err, string(responseData.Body))
	}
	var echo PostmanEchoResponse
	err = json.Unmarshal(responseData.Body, &echo)
	if echo.Args["one"] != request.Params["one"] {
		t.Errorf("Unexpected param one %s", echo.Args["one"])
	}
	if echo.Args["one"] != request.Params["one"] {
		t.Errorf("Unexpected param two %s", echo.Args["two"])
	}
	if echo.Headers["headkey"] != request.Headers["headkey"] {
		t.Errorf("Unexpected header value %s", echo.Headers["headkey"])
	}
	if echo.Headers["user-agent"] != client.UserAgent {
		t.Errorf("Unexpected user-agent %s", client.UserAgent)
	}
	// POST request Test
	uploadFile, err := os.CreateTemp("", "uploadTestFile.txt")
	if err != nil {
		t.Error(err)
	}
	request = RequestData{
		Type:     "POST",
		Url:      "https://postman-echo.com/post",
		BodyData: struct{ BodyData bool }{true},
		Files:    []*os.File{uploadFile},
		Cookies: map[string]string{
			"CookieKey": "1",
		},
	}
	responseData, err = client.QuickRequest(request)
	if err != nil {
		t.Fatal(err, string(responseData.Body))
	}
	t.Error(string(responseData.Body))
	err = json.Unmarshal(responseData.Body, &echo)
	if echo.JSON.(map[string]any)["BodyData"].(bool) != true {
		t.Errorf("Unexpected BodyData %v", echo.JSON)
	}
	if echo.Cookies["CookieKey"] != request.Headers["CookieKey"] {
		t.Errorf("Unexpected cookie value %s", echo.Headers["CookieKey"])
	}
}
