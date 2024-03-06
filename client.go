package HttpClientPool

import (
	"net/http"
	"net/url"
	"time"
)

// Client represents an HTTP client with inbuilt ratelimiting.
//
// This type can be used along with a ClientPool for orchestration of multiple clients.
type Client struct {
	// Client is the underlying HTTP client for making requests.
	*http.Client
	// UserAgent is the user agent string to be set in the client's requests.
	UserAgent   string
	delay       time.Duration
	running     bool
	lastReqTime time.Time
}

// NewClient creates a new HTTP client with optional proxy, user agent, and request delay.
//
// Parameters:
//   - proxy (*url.URL): The proxy URL to be used for the client. Use nil for no proxy.
//   - userAgent (string): The user agent string to be set in the client's requests.
//   - delay (time.Duration): The delay between requests made by the client. Use 0 for no delay.
//
// Returns:
//   - *Client: A pointer to the initialized HTTP client.
func NewClient(proxy *url.URL, userAgent string, delay time.Duration) *Client {
	var httpClient *http.Client
	if proxy != nil {
		// Set proxy
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
		httpClient = &http.Client{Transport: transport}
	} else {
		// No proxy
		httpClient = &http.Client{}
	}

	client := Client{
		Client:    httpClient,
		UserAgent: userAgent,
		delay:     delay,
	}
	return &client
}

// IsRunning returns true if the client is currently running.
//
// This method is used to check if the client is actively processing requests.
//
// Returns:
//   - bool: True if the client is running; otherwise, false.
func (client *Client) IsRunning() bool {
	return client.running
}

// IsAvailable returns true if the client is not currently running or rate-limited.
//
// This method is used to check if the client is in an available state for new requests.
//
// Returns:
//   - bool: True if the client is available; otherwise, false.
func (client *Client) IsAvailable() bool {
	// Check currently running
	if client.running {
		return false
	}
	// Check client ratelimited
	if client.lastReqTime.Add(client.delay).Before(time.Now()) {
		return false
	}
	return true
}
