// Package HttpClientPool allows easy orchestration of an HTTP client pool with built-in rate-limiting capabilities.
//
// Overview:
//
//	This package enables efficient concurrent HTTP requests by managing a pool of individual HTTP clients,
//	each with its own configuration and the ability to apply both per-client and global rate limits.
//
// Usage:
//
//	To use this package, create a ClientPool using the NewClientPool function, specifying the desired
//	client delay, pool delay, optional proxies, and user-agent weights.
//	To make simple requests call ClientPool.QuickRequest() with a RequestData bundle. This will
//	automatically mark it as active and deactivate it when the request is complete.
//
//	For greater flexibility call ClientPool.GetClient() to get an available Client instance and
//	and use it as with a normal http.Client instance. Call Client.SetInactive() when done with
//	the Client to deactivate it.
//
// Example:
//
//	clientPool := HttpClientPool.NewClientPool(time.Millisecond*100, time.Second, nil, nil)
//
// Features:
//   - Dynamic client pool creation with customizable delays.
//   - Rate-limiting for individual clients and the entire pool.
//   - Automatic proxy rotation by ratelimit.
//
// GitHub repository: https://github.com/RootInit/HttpClientPool
package HttpClientPool

import (
	"github.com/RootInit/HttpClientPool/Utils"
	"net/url"
	"time"
)

// ClientPool represents a pool of HTTP clients with easy per client and
// whole pool ratelimiting.
//
// The pool is responsible for managing a collection of HTTP clients, each with its
// own configuration, and a shared delay applied between requests made by clients.
type ClientPool struct {
	// Clients is a slice containing pointers to the clients in the pool.
	Clients []*Client
	delay   time.Duration
}

// NewClientPool creates a pool of HTTP clients for concurrent requests.
//
// Parameters:
//   - clientDelay (time.Duration): Time duration between client requests. Use 0 for no delay.
//   - poolDelay (time.Duration): Time duration between client pool requests. Use 0 for no delay.
//   - proxies ([]*url.URL): List of proxy URLs. Use nil for a single client with no proxy.
//   - userAgents (map[string]float32): Map of user agents with their respective weights.
//
// Returns:
//   - ClientPool: The initialized client pool.
func NewClientPool(clientDelay, poolDelay time.Duration, proxies []*url.URL, userAgents map[string]float32) ClientPool {
	// Create clients
	var clients []*Client
	if proxies == nil {
		// Create single client with no proxy
		client := NewClient(nil, Utils.GetRandomUseragent(userAgents), clientDelay)
		clients = []*Client{client}
	} else {
		// Create client for each proxy
		clients = make([]*Client, len(proxies))
		for idx, proxy := range proxies {
			client := NewClient(proxy, Utils.GetRandomUseragent(userAgents), clientDelay)
			clients[idx] = client
		}
	}
	return ClientPool{
		Clients: clients,
		delay:   poolDelay,
	}
}

// AddClient adds a new HTTP client to the client pool.
//
// This function will accept duplicate clients and add them.
//
// Parameters:
//   - client (*Client): The HTTP client to be added to the pool.
func (pool *ClientPool) AddClient(client *Client) {
	pool.Clients = append(pool.Clients, client)
}

// RemoveClient removes a specific HTTP client from the client pool.
//
// If the client is not in the pool the pool remains unchanged.
// If the client is duplicated in the pool only the first instance
// of the client will be removed.
//
// Parameters:
//   - client (*Client): The HTTP client to be removed from the pool.
func (pool *ClientPool) RemmoveClient(client *Client) {
	for idx, c := range pool.Clients {
		// Compare pointer addresses
		if c == client {
			// Check if last element
			if idx < len(pool.Clients)-1 {
				// Remove from slice
				pool.Clients = append(pool.Clients[:idx], pool.Clients[idx+1:]...)
			} else {
				// Remove the last element
				pool.Clients = pool.Clients[:idx]
			}
			break
		}
	}
}

// SetPoolDelay sets the minimum delay between requests from all clients in the pool.
//
// Parameters:
//   - poolDelay (time.Duration): The new shared delay. Use 0 for no delay.
func (pool *ClientPool) SetPoolDelay(poolDelay time.Duration) {
	pool.delay = poolDelay
}

// SetClientDelay sets the individual delay between requests for each client in the pool.
//
// Parameters:
//   - clientDelay (time.Duration): The new shared delay. Use 0 for no delay.
func (pool *ClientPool) SetClientDelay(clientDelay time.Duration) {
	for _, client := range pool.Clients {
		client.SetDelay(clientDelay)
	}
}

// GetClient returns an available HTTP client from the pool.
// The client is set as active and the lastReqTime is set to time.Now.
//
// This method blocks until a client becomes available in the pool.
//
// Returns:
//   - *Client: A pointer to the available HTTP client.
func (pool *ClientPool) GetClient() *Client {
	// Calculate time since last request
	var lastReqTime time.Time
	for _, client := range pool.Clients {
		clientReqTime := client.GetRequestTime()
		if clientReqTime.After(lastReqTime) {
			lastReqTime = clientReqTime
		}
	}
	lastReqDelta := time.Now().Sub(lastReqTime)
	if lastReqDelta < pool.delay {
		// Wait until pool delay elapsed
		time.Sleep(pool.delay - lastReqDelta)
	}
	for {
		for _, client := range pool.Clients {
			if client.IsAvailable() {
				client.SetActive()
				return client
			}
		}
		time.Sleep(time.Millisecond * 1)
	}
}

// QuickRequest is a convenience function which fetches a Client
// with pool.GetClient and passes the RequestData to client.QuickRequest
//
// Parameters:
//   - reqData (RequestData): The RequestData struct containing HTTP request data.
//
// Returns:
//   - ResponseData: A ResponseData struct containing HTTP response data.
//   - error: An error, if any, encountered during the HTTP request.
func (pool *ClientPool) QuickRequest(reqData RequestData) (ResponseData, error) {
	client := pool.GetClient()
	return client.QuickRequest(reqData)
}

// Done blocks until all clients in the pool are inactive.
//
// This method ensures that all active clients finish their ongoing requests
// before allowing the program to proceed.
func (pool *ClientPool) Done() {
	for {
		done := true
		for _, client := range pool.Clients {
			if client.IsRunning() {
				done = false
				break
			}
		}
		if done {
			return
		}
		time.Sleep(10)
	}
}
