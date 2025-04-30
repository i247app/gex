// cmd/app1/demo_run.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cmd/app1/client" // Import the client package
)

func main() {
	//  Get the App1 service base URL from an environment variable.
	//  This is a common way to configure the client in a real application.
	app1BaseURL := os.Getenv("APP1_BASE_URL")
	if app1BaseURL == "" {
		log.Fatal("APP1_BASE_URL environment variable not set")
	}

	// Create a new client.
	c := client.NewClient(app1BaseURL)

	//  Create a background context.  Good for simple tests.  For more
	//  complex scenarios, you might use a context with a timeout or
	//  cancellation.
	ctx := context.Background()

	// =========================================================================
	//  Example Test Calls
	// =========================================================================

	fmt.Println("--- Echo Test ---")
	echoMsg := "Hello, App1 Service!"
	echoResponse, err := c.Echo(ctx, echoMsg)
	if err != nil {
		log.Printf("Echo test failed: %v", err)
	} else {
		fmt.Printf("Echo Response: %s\n", echoResponse)
	}

	fmt.Println("\n--- Login Test ---")
	//  Replace with actual test credentials.  Do NOT hardcode real passwords.
	loginToken, err := c.Login(ctx, "testuser", "password")
	if err != nil {
		log.Printf("Login test failed: %v", err)
	} else {
		fmt.Printf("Login Token: %s\n", loginToken)
	}

	fmt.Println("\n--- Logout Test ---")
	err = c.Logout(ctx)
	if err != nil {
		log.Printf("Logout test failed: %v", err)
	} else {
		fmt.Printf("Logout successful\n")
	}

	fmt.Println("\n--- Time Test ---")
	timeResponse, err := c.Time(ctx)
	if err != nil {
		log.Printf("Time test failed: %v", err)
	} else {
		fmt.Printf("Time Response: %s\n", timeResponse)
	}

	fmt.Println("\n--- WhoAmI Test ---")
	whoAmIResponse, err := c.WhoAmI(ctx)
	if err != nil {
		log.Printf("WhoAmI test failed: %v", err)
	} else {
		fmt.Printf("WhoAmI Response: %s\n", whoAmIResponse)
	}
}
