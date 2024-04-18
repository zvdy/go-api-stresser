package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
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
		for i := 0; i < *iterations; i++ {
			go func() {
				_, err := client.Do(req)
				if err != nil {
					fmt.Println("Error in stress test iteration:", err)
				}
			}()
		}
		time.Sleep(time.Duration(*duration) * time.Second)
	}
}

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: api-tester [options]")
		fmt.Println("Flags:")
		flag.PrintDefaults()
	}
}
