package Utils

import (
	"math"
	"testing"
	"os"
)

func TestUrlsFromFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_urls.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer tempFile.Close()
	urlsToWrite := []string{
		"http://example.com/testpath",
		"https://example.org",
		"ftp://example.net",
	}
	for _, u := range urlsToWrite {
		_, err := tempFile.WriteString(u + "\n")
		if err != nil {
			t.Fatal(err)
		}
	}
  // Read back from the file
	urls, err := UrlsFromFile(tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	// Check if the result matches the expected URLs
	if len(urls) != len(urlsToWrite) {
		t.Fatalf("Expected %d URLs, got %d", len(urlsToWrite), len(urls))
	}
	for i, expectedURL := range urlsToWrite {
		if urls[i].String() != expectedURL {
			t.Errorf("Mismatch at index %d. Expected: %s, Got: %s", i, expectedURL, urls[i].String())
		}
	}
}

func TestRandom(t *testing.T) {
	testWeights := []float32{10, 30, 60}
	// Large number required for predictible result
	const samples = 100000
	resultMap := make(map[int]int, len(testWeights))
	for i := 0; i < samples; i++ {
		resultIdx := weightedRandom(testWeights)
		resultMap[resultIdx] += 1
	}
	for idx, count := range resultMap {
    probability := int(math.Round(float64(testWeights[idx])))
		// Round to two sigfigs
		roundedCount := int(math.Round(float64(count) / (samples / 100)))
		if !(probability-1 <= roundedCount) || !(roundedCount <= probability+1) {
			t.Errorf("Expected range %d-%d got %d", probability-1, probability+1, roundedCount)
		}
	}
}


