package HttpClientPool

// Note: Tests use unrealistically low delay for fast tests.

import (
	"testing"
	"time"

	"github.com/RootInit/HttpClientPool/Utils"
)

// Tests a client with no pool
func TestBareClientRatelimiting(t *testing.T) {
	// Create Client
	client := NewClient(nil, "HttpClient", 0)
  // Test active/inactive blocking
	client.SetActive()
	if client.IsAvailable() {
		t.Error("Client should not be available yet.")
	}
	client.SetInactive()
	if !client.IsAvailable() {
		t.Error("Client should be available but is not.")
	}
  // Set delay blocking
  client.SetDelay(10*time.Millisecond)
  client.SetActive()
  client.SetInactive()
	if client.IsAvailable() {
		t.Error("Client should not be available yet.")
	}
	time.Sleep(10 * time.Millisecond)
	if !client.IsAvailable() {
		t.Error("Client should be available but is not.")
	}
}

// Test a single client pool
func TestSingleClientPoolRatelimiting(t *testing.T) {
	// Create pool with no client or pool delay
	pool := NewClientPool(0, 0, nil, nil)
	// Get client from pool
  tic := time.Now()
  client := pool.GetClient()
  client.SetInactive()
  client = pool.GetClient()
  // Should have taken 0 (rounded down) milliseconds
  timeSpent := time.Now().Sub(tic)/time.Millisecond
	if timeSpent != 0 {
    t.Errorf("GetClient took an unexpected amount of time (%d)", timeSpent)
  }
  client.SetInactive()
	// Test client delay
	clientDelay := Utils.MillisecondToDuration(10)
  pool.SetClientDelay(clientDelay)
  tic = time.Now()
  client.SetActive() // Reset lastReqTime
  client.SetInactive()
  client = pool.GetClient()
  timeSpent = time.Now().Sub(tic)/time.Millisecond
	if timeSpent != 10 {
    t.Errorf("GetClient took an unexpected amount of time (%d)", timeSpent)
  }
	// Repeat previous test with pool delay
	poolDelay := Utils.MillisecondToDuration(20)
  pool.SetPoolDelay(poolDelay)
  tic = time.Now()
  client.SetActive() // Reset lastReqTime
  client.SetInactive()
  client = pool.GetClient()
  timeSpent = time.Now().Sub(tic)/time.Millisecond
	if timeSpent != 20 {
    t.Errorf("GetClient took an unexpected amount of time (%d)", timeSpent)
  }
}

// Test a multi-client pool
func TestMultiClientPoolRatelimiting(t *testing.T) {
  //TODO
}

// Tests creating an initially empty pool then adding and removing clients
func TestAddRemoveClients(t *testing.T) {
	// Create pool with default single client
	pool := NewClientPool(0, 0, nil, map[string]float32{"HttpPoolClient": 1})
	if len(pool.Clients) != 1 {
		t.Errorf("Incorrect pool size. Expected 1 got %d", len(pool.Clients))
	}
	// Remove default client
	client := pool.GetClient()
	client.SetInactive()
	pool.RemmoveClient(client)
	if len(pool.Clients) != 0 {
		t.Errorf("Incorrect pool size. Expected 0 got %d", len(pool.Clients))
	}
	// Add new client three times
	client = NewClient(nil, "TestClient", 0)
	for i := 0; i < 3; i++ {
		pool.AddClient(client)
	}
	if len(pool.Clients) != 3 {
		t.Errorf("Incorrect pool size. Expected 3 got %d", len(pool.Clients))
	}
	// Remove client (only one instance should be removed)
	pool.RemmoveClient(client)
	if len(pool.Clients) != 2 {
		t.Errorf("Incorrect pool size. Expected 2 got %d", len(pool.Clients))
	}
	// Remove all clients
	for _, client := range pool.Clients {
    pool.RemmoveClient(client)
	}
  if len(pool.Clients) != 0 {
		t.Errorf("Incorrect pool size. Expected 0 got %d", len(pool.Clients))
	}
}
