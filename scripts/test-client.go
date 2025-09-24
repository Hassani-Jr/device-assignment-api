package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Simple test client to demonstrate mTLS authentication
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test-client.go <command>")
		fmt.Println("Commands:")
		fmt.Println("  auth         - Test device authentication")
		fmt.Println("  generate-jwt - Generate a test JWT token")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "auth":
		testDeviceAuth()
	case "generate-jwt":
		generateTestJWT()
	default:
		fmt.Println("Unknown command:", command)
		os.Exit(1)
	}
}

func testDeviceAuth() {
	// Load client certificate
	cert, err := tls.LoadX509KeyPair("./certs/client.crt", "./certs/client.key")
	if err != nil {
		log.Fatal("Failed to load client certificate:", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile("./certs/ca.crt")
	if err != nil {
		log.Fatal("Failed to read CA certificate:", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Configure TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   "localhost",
	}

	// Create HTTP client with mTLS
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// Make request to authenticate endpoint
	resp, err := client.Post("https://localhost:8443/api/v1/devices/authenticate", "application/json", nil)
	if err != nil {
		log.Fatal("Failed to make request:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to read response:", err)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response: %s\n", string(body))

	if resp.StatusCode == http.StatusOK {
		var device map[string]interface{}
		if err := json.Unmarshal(body, &device); err == nil {
			fmt.Printf("Device ID: %s\n", device["id"])
			fmt.Printf("Serial Number: %s\n", device["certificate_serial_number"])
		}
	}
}

func generateTestJWT() {
	// This would typically be done by a separate authentication service
	// For demo purposes, we'll show what a JWT token would look like
	fmt.Println("In a real implementation, you would:")
	fmt.Println("1. Authenticate with your user management system")
	fmt.Println("2. Receive a JWT token")
	fmt.Println("3. Use that token for API calls")
	fmt.Println()
	fmt.Println("Example JWT usage:")
	fmt.Println("curl -H 'Authorization: Bearer <your-jwt-token>' \\")
	fmt.Println("     https://localhost:8443/api/v1/users/me/devices")
}
