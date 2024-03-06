/*
Provides a ClientPool that wraps http.Clients to allow easy ratelimiting of
several independant http.Clients with individual ratelimits
*/
package HttpClientPool

import (
	"HttpPool/Utils"
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
		for _, proxy := range proxies {
			client := NewClient(proxy, Utils.GetRandomUseragent(userAgents), clientDelay)
			clients = append(clients, client)
		}
	}
	return ClientPool{
		Clients: clients,
		delay:   poolDelay,
	}
}

// AddClient adds a new HTTP client to the client pool.
//
// Parameters:
//   - client (*Client): The HTTP client to be added to the pool.
func (pool ClientPool) AddClient(client *Client) {
	pool.Clients = append(pool.Clients, client)
}

// SetPoolDelay sets the minimum delay between requests from all clients in the pool.
//
// Parameters:
//   - poolDelay (time.Duration): The new shared delay. Use 0 for no delay.
func (pool ClientPool) SetPoolDelay(poolDelay time.Duration) {
	pool.delay = poolDelay
}

// SetClientDelay sets the individual delay between requests for each client in the pool.
//
// Parameters:
//   - clientDelay (time.Duration): The new shared delay. Use 0 for no delay.
func (pool ClientPool) SetClientDelay(clientDelay time.Duration) {
	for _, client := range pool.Clients {
		client.delay = clientDelay
	}
}

// GetClient returns an available HTTP client from the pool.
//
// This method blocks until a client becomes available in the pool.
//
// Returns:
//   - *Client: A pointer to the available HTTP client.
func (pool ClientPool) GetClient() *Client {
	for {
		// Get most recent request time
		var lastReqTime time.Time
		for _, client := range pool.Clients {
			if client.lastReqTime.After(lastReqTime) {
				lastReqTime = client.lastReqTime
        client.SetActive()
			}
		}
		// Check pool not ratelimited
		if !lastReqTime.Add(pool.delay).After(time.Now()) {
			for _, client := range pool.Clients {
				if client.IsAvailable() {
					return client
				}
			}
		}
		time.Sleep(time.Millisecond * 10)
	}
}

// Done blocks until all clients in the pool are inactive.
//
// This method ensures that all active clients finish their ongoing requests
// before allowing the program to proceed.
func (pool ClientPool) Done() {
	for {
		done := true
		for _, client := range pool.Clients {
			if client.running {
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