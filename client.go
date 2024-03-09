package HttpClientPool

import (
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Client represents an HTTP client with inbuilt ratelimiting.
//
// This type can be used along with a ClientPool for orchestration of multiple clients.
type Client struct {
	// Client is the underlying HTTP client for making requests.
	*http.Client
	// userAgent is the user agent string to be set in the client's requests.
	userAgent   string
	delay       time.Duration
	running     bool
	lastReqTime time.Time
	mu          sync.Mutex
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
		userAgent: userAgent,
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
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.running
}

// SetActive marks the HTTP client as active and updates the lastReqTime.
func (client *Client) SetActive() {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.running = true
	client.lastReqTime = time.Now()
}

// SetInactive marks the HTTP client as inactive.
func (client *Client) SetInactive() {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.running = false
}

// SetDelay ets the clients delay
//
// Parameters:
//   - delay (time.Duration): The duration of the new delay
func (client *Client) SetDelay(delay time.Duration) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.delay = delay
}

// GetDelay returns the clients delay
//
// Returns:
//   - time.Duration: The duration of the clients delay
func (client *Client) GetDelay() time.Duration {
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.delay
}

// IsAvailable returns true if the client is not currently running or rate-limited.
//
// This method is used to check if the client is in an available state for new requests.
//
// Returns:
//   - bool: True if the client is available; otherwise, false.
func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	// Check currently running
	if client.running {
		return false
	}
	// Check client ratelimited
	if client.lastReqTime.Add(client.delay).After(time.Now()) {
		return false
	}
	return true
}

// GetUserAgent sets the Clients user-agent
//
// Parameters:
//   - userAgent (string): the userAgent to set
func (client *Client) SetUserAgent(userAgent string) {
	client.mu.Lock()
	defer client.mu.Unlock()
  client.userAgent = userAgent
}
// GetUserAgent returns the Clients user-agent
//
// Returns:
//   - string: the client.userAgent value 
func (client *Client) GetUserAgent() string {
	client.mu.Lock()
	defer client.mu.Unlock()
  return client.userAgent
}
// GetRequestTime returns the last request time
//
// Returns:
//   - time.Time: the client.lastReqTime value 
func (client *Client) GetRequestTime() time.Time {
	client.mu.Lock()
	defer client.mu.Unlock()
  return client.lastReqTime
}
