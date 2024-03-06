
<h1 align="center">HttpClientPool</h1>

<p align="center">
  <b>Easy orchestration of an HTTP client pool with built-in rate-limiting capabilities.</b>
</p>

---

## Overview

This package facilitates efficient concurrent HTTP requests by managing a pool of individual HTTP clients, each with its configuration and the ability to apply both per-client and global rate limits.

## Usage

To use this package, create a `ClientPool` using the `NewClientPool` function, specifying the desired client delay, pool delay, optional proxies, and user-agent weights. For simple requests, use `ClientPool.QuickRequest()` with a `RequestData` bundle. This automatically marks it as active and deactivates it when the request is complete.

For greater flexibility, use `ClientPool.GetClient()` to get an available `Client` instance and use it as with a normal `http.Client` instance. Call `Client.SetInactive()` when done with the client to deactivate it.

## Example

```go
clientPool := HttpClientPool.NewClientPool(time.Millisecond*100, time.Second, nil, nil)

