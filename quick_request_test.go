package HttpClientPool

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"
	"sync"
	"testing"
	"time"
  "context"
)

// Runs a battery of various web requests
func TestQuickRequest(t *testing.T) {
	// Create Client
	client := NewClient(nil, "HttpClient", 0)
	// Start echo webserver
	serverDone := &sync.WaitGroup{}
	server := startEchoWebserver(serverDone)
  // Run Requests
	getRequestTest(client, t)
	postJsonRequestTest(client, t)
	postFormRequestTest(client, t)
	// Stop echo webserver
	shutdownCtx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err := server.Shutdown(shutdownCtx)
	if err != nil {
		t.Error(err)
	}

	serverDone.Wait()
}

// RequestData represents the data to be returned by the echo webserver
type EchoData struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Params  map[string][]string `json:"params"`
	Body    interface{}         `json:"body"`
	Cookies map[string]string   `json:"cookies"`
}

func startEchoWebserver(wg *sync.WaitGroup) *http.Server {

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			print(err)
			return
		}
		// Parse request cookies
		cookies := make(map[string]string)
		for _, cookie := range r.Cookies() {
			cookies[cookie.Name] = cookie.Value
		}
		requestData := EchoData{
			Method:  r.Method,
			URL:     r.URL.Path,
			Headers: r.Header,
			Params:  r.URL.Query(),
			Body:    string(body),
			Cookies: cookies,
		}
		// Return the data
		responseJSON, err := json.MarshalIndent(requestData, "", "  ")
		if err != nil {
			var errMsg = "Error creating JSON response" + err.Error()
			http.Error(w, errMsg, http.StatusInternalServerError)
			print(errMsg)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})
	server := &http.Server{Addr: ":8080", Handler: mux}
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Start the web server on port 8080
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			println("Error running webserver:", err)
		}
	}()
	// Give time for webserver to start
	time.Sleep(time.Millisecond * 100)
	return server
}

func getRequestTest(client *Client, t *testing.T) {
	request := RequestData{
		Type: "GET",
		Url:  "http://127.0.0.1:8080",
		Params: map[string][]string{
			"one": {"1"},
			"two": {"2"},
		},
		Headers: map[string][]string{
			"Headkey": {"1"},
		},
		Cookies: map[string]string{
			"CookieKey": "1",
		},
	}
	responseData, err := client.QuickRequest(request)
	if err != nil {
		t.Fatal(err, string(responseData.Body))
	}
	var echo EchoData
	err = json.Unmarshal(responseData.Body, &echo)

	// Check param "one"
	if _, exists := echo.Params["one"]; !exists {
		t.Error("Key \"one\" not in echo params")
		t.Log(jsonFmt(echo.Params))
		return
	}
	if echo.Params["one"][0] != request.Params["one"][0] {
		t.Errorf("Unexpected param one %s", echo.Params["one"])
	}
	// Check param "two"
	if _, exists := echo.Params["two"]; !exists {
		t.Error("Key \"two\" not in echo params")
		t.Log(jsonFmt(echo.Params))
		return
	}
	if echo.Params["two"][0] != request.Params["two"][0] {
		t.Errorf("Unexpected param two %s", echo.Params["two"])
	}
	// Check header "Headkey"
	if _, exists := echo.Headers["Headkey"]; !exists {
		t.Error("Key \"Headkey\" not in echo headers")
		t.Log(jsonFmt(echo.Headers))
		return
	}
	if echo.Headers["Headkey"][0] != request.Headers["Headkey"][0] {
		t.Errorf("Unexpected header value %s", echo.Headers["Headkey"])
	}
	// Check header "User-Agent" and ensure no duplicate key
	if size := len(echo.Headers["User-Agent"]); size != 1 {
		t.Error("Key \"User-Agent\" incorrect size in echo headers")
		t.Log(jsonFmt(echo.Headers))
		return
	}
	if echo.Headers["User-Agent"][0] != client.GetUserAgent() {
		t.Errorf("Unexpected user-agent %s", echo.Headers["User-Agent"][0])
	}
	// Check cookie "CookieKey"
	if _, exists := echo.Cookies["CookieKey"]; !exists {
		t.Error("Key \"CookieKey\" not in echo cookies")
		t.Log(jsonFmt(echo.Cookies))
		return
	}
	if echo.Cookies["CookieKey"] != request.Cookies["CookieKey"] {
		t.Errorf("Unexpected cookie value %s", echo.Headers["CookieKey"])
	}
}

func postJsonRequestTest(client *Client, t *testing.T) {
	request := RequestData{
		Type:     "POST",
		Url:      "http://127.0.0.1:8080",
		JsonData: struct{ BodyData bool }{true},
	}
	responseData, err := client.QuickRequest(request)
	if err != nil {
		t.Fatal(err, jsonFmt(responseData.Body))
	}
	var echo EchoData
	err = json.Unmarshal(responseData.Body, &echo)
	if echo.Body.(string) != "{\"BodyData\":true}" {
		t.Errorf("Unexpected BodyData %v", jsonFmt(echo.Body))
	}
}

func postFormRequestTest(client *Client, t *testing.T) {
	// Create test file
	uploadFile, err := os.CreateTemp("", "uploadTestFile.txt")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(uploadFile.Name())
	_, err = uploadFile.WriteString("testdata")
	if err != nil {
		t.Error(err)
	}
	request := RequestData{
		Type:      "POST",
		Url:       "http://127.0.0.1:8080",
		FormData:  map[string]string{"Field1": "true"},
		FormFiles: map[string]*os.File{"attachment": uploadFile},
	}
	// Generate expected echo string
	reader, err := formDataReader(request.FormData, request.FormFiles)
	if err != nil {
		t.Error(err)
	}
	expectedResult, err := io.ReadAll(reader)
	if err != nil {
		t.Error(err)
	}
	responseData, err := client.QuickRequest(request)
	if err != nil {
		t.Fatal(err, jsonFmt(responseData.Body))
	}
	var echo EchoData
	err = json.Unmarshal(responseData.Body, &echo)
	// Check body
	re := regexp.MustCompile(`--([a-fA-F0-9]+)`)
	echoBody := re.ReplaceAllString(echo.Body.(string), "")
	expectedBody := re.ReplaceAllString(string(expectedResult), "")
	if echoBody != expectedBody {
		t.Errorf("Unexpected BodyData\n%v\nExpected:\n%v", jsonFmt(echoBody), jsonFmt(expectedBody))
	}
}

func jsonFmt(o interface{}) string {
	responseJSON, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		panic("Error marshaling object")
	}
	return string(responseJSON)
}
