package Utils

import (
	"bufio"
	"math/rand"
	"net/url"
	"strings"
	"time"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"os"
)

// MillisecondToDuration converts a time duration in milliseconds to a time.Duration.
//
// The ms argument must be able to be cast as a float64. Negative values will be rounded to 0.
// The function supports int, float32, and float64 types for the ms argument.
//
// Parameters:
//   - ms (interface{}): The time duration in milliseconds, which can be an int, float32, or float64.
//
// Returns:
//   - time.Duration: The equivalent time.Duration value.
func MillisecondToDuration(ms interface{}) time.Duration {
  var msValue float64
	switch ms.(type) {
	case int:
		msValue = float64(ms.(int))
	case float32:
		msValue = float64(ms.(float32))
	case float64:
		msValue = ms.(float64)
	default:
    panic("Expected int or float ms value")
	}
	return time.Millisecond * time.Duration(msValue)
}

// RpsToDuration converts a rate per second (rps) to a time.Duration delay.
//
// If the rate per second is less than or equal to 0, the function returns a time.Duration of 0.
//
// Parameters:
//   - rps (float32): The rate per second for which to calculate the time duration delay.
//
// Returns:
//   - time.Duration: The calculated time duration delay based on the given rate per second.
func RpsToDuration(rps float32) time.Duration {
	if rps <= 0 {
		return time.Duration(0)
	} else {
		return MillisecondToDuration(1000 / rps)
	}
}

// UrlsFromFile reads a file containing URLs (one per line) and returns a slice of parsed URL objects.
//
// The function reads the content of the specified file, expecting one URL per line. It trims whitespace
// from each line and parses it into a URL. The resulting parsed URLs are returned as a slice.
//
// Parameters:
//   - filename (string): The name of the file containing URLs.
//
// Returns:
//   - []*url.URL: A slice of parsed URL objects.
//   - error: An error, if any, encountered during file reading or URL parsing.
func UrlsFromFile(filename string) ([]*url.URL, error) {
	var urls []*url.URL
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			parsedUrl, err := url.Parse(line)
			if err != nil {
				return nil, err
			}
			urls = append(urls, parsedUrl)
		}
	}
	err = scanner.Err()
	return urls, err
}

// GetRandomUseragent returns a weighted random user agent from a map of 
// user agents with associated weights.
//
// If the input map is nil, falls back default user agents with predefined weights.
// A source for current user-agents is https://www.useragents.me/ 
//
// Parameters:
//   - useragents (map[string]float32): A map of user agents with associated weights.
//
// Returns:
//   - string: The selected user agent based on weighted random selection.
func GetRandomUseragent(useragents map[string]float32) string {
	// Set fallback useragents
	if useragents == nil {
		useragents = map[string]float32{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.1": 49.09,
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.1": 14.33,
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.3":       13.41,
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.3":       8.54,
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/117.":                                      3.66,
		}
	}
	mapSize := len(useragents)
	// Sum weights and convert into two slices
	var uas = make([]string, 0, mapSize)
	var weights = make([]float32, 0, mapSize)
	for ua, weight := range useragents {
		uas = append(uas, ua)
		weights = append(weights, weight)
	}
	return uas[weightedRandom(weights)]
}

// weightedRandom returns the index of a random choice from an array of weights.
//
// Intended to be used with a second array of values to be accessed using the index.
// Weights do not need to total to 100.
//
// Parameters:
//   - weights ([]float32): An array of weights representing the likelihood of each choice.
//
// Returns:
//   - int: The index of the chosen element based on the weighted random selection.
func weightedRandom(weights []float32) int {
	// Sum weights and convert into two slices
	var total float32 = 0
	var cumWeights = make([]float32, 0, len(weights))
	for _, weight := range weights {
		total += weight
		cumWeights = append(cumWeights, total)
	}
	// Get random value between 0 and total
	rnd := rand.Float32() * total
	// Itterate through slices backwards
	for idx, weight := range cumWeights {
		if rnd <= weight {
			return idx
		}
	}
	panic("This is impossible")
}


// JsonDataReader converts a Go data structure into a JSON-formatted io.Reader.
//
// Parameters:
//   - data (interface{}): The Go data structure to be converted to JSON.
//
// Returns:
//   - io.Reader: An io.Reader containing the JSON-encoded data.
//   - error: An error, if any, encountered during the JSON encoding process.
func JsonDataReader(data interface{}) (io.Reader, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	bodyData := bytes.NewBuffer(jsonData)
	return bodyData, nil
}

// FormDataReader creates a multipart/form-data io.Reader from a map of key-value pairs
// and a map of files.
//
// Parameters:
//   - data (map[string]string): A map of string key-value pairs representing form fields.
//   - files (map[string]*os.File): A map of files to be included in the request body.
//
// Returns:
//   - io.Reader: An io.Reader containing the multipart/form-data request body.
//   - error: An error, if any, encountered during the construction of the request body.
func FormDataReader(data map[string]string, files map[string]*os.File) (io.Reader, error) {
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
