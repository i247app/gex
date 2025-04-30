// cmd/app1/demo/demo.go
package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client is a struct that holds the base URL of the App1 service.
type Client struct {
	baseURL *url.URL
	client  *http.Client // Use a custom http.Client if needed for timeouts, etc.
}

// NewClient creates a new Client.  It takes the base URL of the App1
// service as a parameter.  It's good practice to parse the URL here
// to catch any errors early.
func NewClient(baseURL string) *Client {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		//  In a real application, you might want to return the error
		//  instead of panicking.  However, for a simple client,
		//  panicking in the constructor is often acceptable.  The
		//  caller should provide a valid URL.
		panic(fmt.Sprintf("invalid base URL: %v", err))
	}
	return &Client{
		baseURL: parsedURL,
		client:  &http.Client{}, //  Use the default client.
	}
}

// request is a helper method that makes an HTTP request to the App1 service.
// It handles constructing the full URL, setting the method, adding the body,
// and handling the response.  This simplifies the other methods.
func (c *Client) request(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	// Construct the full URL.  Make sure to use PathJoin, which correctly
	// handles slashes.
	fullURL := c.baseURL.JoinPath(path).String()

	var reqBody io.Reader
	if body != nil {
		//  Use json.Marshal to convert the body to JSON.  This works for
		//  most API requests.  If you need to send form data or other
		//  formats, you'll need to add a parameter to the request
		//  method to specify the content type and use a different
		//  encoding function.
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = strings.NewReader(string(jsonBody))
	}

	// Create the HTTP request.
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the Content-Type header for JSON.
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Make the HTTP request.  Use the client from the Client struct.
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	//  It's the caller's responsibility to close the response body.
	//  We could use a defer statement here, but it's more common to
	//  leave it to the caller.

	// Check the response status code.  Handle common error cases.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read the response body so we can include it in the error message.
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("request failed with status %s, and failed to read response body: %w", resp.Status, readErr)
		}
		resp.Body.Close() // Close here because we are returning
		return nil, fmt.Errorf("request failed with status %s, body: %s", resp.Status, string(bodyBytes))
	}

	return resp, nil
}

// echoResponse is a struct to hold the response from the /echo endpoint.
type echoResponse struct {
	Message string `json:"message"`
}

// Echo sends a message to the /echo endpoint and returns the response.
func (c *Client) Echo(ctx context.Context, msg string) (string, error) {
	//  Use the helper method to make the request.
	resp, err := c.request(ctx, http.MethodPost, "/echo", map[string]string{"message": msg})
	if err != nil {
		return "", err //  The error from c.request already has context.
	}
	defer resp.Body.Close()

	// Read the entire response body.  Use io.ReadAll.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal the JSON response into the echoResponse struct.
	var echoResp echoResponse
	if err := json.Unmarshal(body, &echoResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return echoResp.Message, nil
}

// loginRequest is a struct for the request body of the /login endpoint.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// loginResponse is a struct for the response body of the /login endpoint.
type loginResponse struct {
	Token string `json:"token"`
}

// Login sends a username and password to the /login endpoint and returns the token.
func (c *Client) Login(ctx context.Context, username, password string) (string, error) {
	reqBody := loginRequest{Username: username, Password: password}
	resp, err := c.request(ctx, http.MethodPost, "/login", reqBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var loginResp loginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return loginResp.Token, nil
}

// Logout sends a request to the /logout endpoint.
func (c *Client) Logout(ctx context.Context) error {
	resp, err := c.request(ctx, http.MethodPost, "/logout", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//  The /logout endpoint may return a 200 with an empty body, or
	//  a 204 No Content.  We need to handle both cases.
	if resp.StatusCode == http.StatusNoContent {
		return nil // Success
	}

	//  If we get here, it was a 200, so we should read the body.
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	return nil
}

// timeResponse is a struct for the response body of the /time endpoint.
type timeResponse struct {
	Time string `json:"time"`
}

// Time gets the current time from the /time endpoint.
func (c *Client) Time(ctx context.Context) (string, error) {
	resp, err := c.request(ctx, http.MethodGet, "/time", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var timeResp timeResponse
	if err := json.Unmarshal(body, &timeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return timeResp.Time, nil
}

// whoAmIResponse is a struct for the response body of the /whoami endpoint.
type whoAmIResponse struct {
	UserID string `json:"userID"`
}

// WhoAmI gets the current user's ID from the /whoami endpoint.
func (c *Client) WhoAmI(ctx context.Context) (string, error) {
	resp, err := c.request(ctx, http.MethodGet, "/whoami", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var whoAmIResp whoAmIResponse
	if err := json.Unmarshal(body, &whoAmIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return whoAmIResp.UserID, nil
}
