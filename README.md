
<h1 align="center">HttpClientPool</h1>

<p align="center">
  <b>Easy orchestration of an HTTP client pool with built-in rate-limiting capabilities.</b>
</p>

---

## Disclaimer

This was made purely for personal use. While everything seems to work in my use cases and passes tests unexpected behavior may occur.

## Overview

This package facilitates efficient concurrent HTTP requests by managing a pool of individual HTTP clients, each with its configuration and the ability to apply both per-client and global rate limits.

## Usage

To use this package, create a `ClientPool` using the `NewClientPool` function, specifying the desired client delay, pool delay, optional proxies, and user-agent weights. For simple requests, use `ClientPool.QuickRequest()` with a `RequestData` bundle. This automatically marks it as active and deactivates it when the request is complete.

For greater flexibility, use `ClientPool.GetClient()` to get an available `Client` instance and use it as with a normal `http.Client` instance. Call `Client.SetInactive()` when done with the client to deactivate it.

## Example

```go
proxies := Utils.UrlsFromFile("proxies.txt")
clientDelay := time.Millisecond * 500//  Two requests per second
poolDelay := Utils.RpsToDuration(25)//25 requests per second
pool := HttpClientPool.NewClientPool(clientDelay, poolDelay, proxies, nil)
// Fetch first 1000 pages
for i:=0;i<1000;i++ {
	request := RequestData{
		Type: "GET",
		Url:  "http://api.com/getByIndex",
		Params: map[string][]string{
			"Index": {i},
		},
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Cookies: map[string]string{
			"Api-Key": "5318008",
		},
	}
	responseData, err := client.QuickRequest(request)
	if err != nil {
		t.Fatal(err, string(responseData.Body))
	}
    // Do something with responseData
}
```

