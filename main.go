package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

// HttpRequest struct to store request parameters
type HttpRequest struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Headers map[string]string      `json:"config"` // Renamed for clarity
	Body    map[string]interface{} `json:"body"`
}

var (
	configFile = flag.String("config", "config.json", "Path to the configuration file")
	iterations = flag.Int("iterations", 1, "Number of iterations for stress test")
	duration   = flag.Int("duration", 10, "Duration of stress test in seconds")
)

func main() {
	flag.Parse()

	if *configFile == "" {
		fmt.Println("Error: Configuration file path is required. Use -h or --help for usage.")
		os.Exit(1)
	}

	if *iterations <= 0 {
		fmt.Println("Error: Number of iterations must be positive. Use -h or --help for usage.")
		os.Exit(1)
	}

	if *duration <= 0 {
		fmt.Println("Error: Stress test duration must be positive. Use -h or --help for usage.")
		os.Exit(1)
	}

	// Read API request configuration from a JSON file
	data, err := os.Open(*configFile)
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return
	}
	defer data.Close()

	var request HttpRequest
	decoder := json.NewDecoder(data)
	err = decoder.Decode(&request)
	if err != nil {
		fmt.Println("Error parsing config JSON:", err)
		return
	}

	// Build the HTTP request
	client := &http.Client{}
	req, err := http.NewRequest(request.Method, request.URL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Add headers to the request
	for key, value := range request.Headers {
		req.Header.Add(key, value)
	}

	// Add body to the request (optional)
	if request.Body != nil {
		jsonData, err := json.Marshal(request.Body)
		if err != nil {
			fmt.Println("Error marshalling body:", err)
			return
		}
		req.Body = io.NopCloser(bytes.NewReader(jsonData))
	}

	// Send the request
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Process the response
	elapsed := time.Since(start)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Response Time:", elapsed)
	fmt.Println("Response Body:", string(body))

	// Perform stress test (optional)
	if *iterations > 1 {
		fmt.Println("Running stress test...")

		startTime := time.Now()
		totalTime := time.Duration(*duration) * time.Second

		// Create a new progress bar with a mutex for synchronized updates
		bar := progressbar.NewOptions(
			100, // Set total progress to 100 for percentage representation
			progressbar.OptionSetWriter(os.Stdout),
			progressbar.OptionSetDescription("[Performing Stress Test]"),
			progressbar.OptionSetWidth(30),
		)
		var barMutex sync.Mutex

		go func() {
			for {
				elapsed := time.Since(startTime)
				progress := int(float64(elapsed) / float64(totalTime) * 100)

				barMutex.Lock()
				bar.Set(progress) // Update progress bar based on elapsed time
				barMutex.Unlock()

				// Check if duration has elapsed
				if elapsed >= totalTime {
					break
				}
				time.Sleep(time.Second) // Sleep for a short duration to avoid busy waiting
			}
			bar.Finish()
		}()

		time.Sleep(totalTime) // Wait for the entire duration
	}
}
