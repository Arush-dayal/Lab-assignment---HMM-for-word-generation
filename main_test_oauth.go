package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	baseURL := "http://localhost:8080"

	// Authorize
	authURL := fmt.Sprintf("%s/auth?redirect_uri=%s&state=%s",
		baseURL,
		url.QueryEscape("http://localhost/callback"),
		"abc123",
	)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(authURL)
	if err != nil {
		log.Fatalf("Auth request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		log.Fatalf("Expected 302 redirect, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	fmt.Println("Redirect location:", location)

	// Parse redirect to get code & state
	redirectURL, err := url.Parse(location)
	if err != nil {
		log.Fatalf("Invalid redirect: %v", err)
	}

	code := redirectURL.Query().Get("code")
	state := redirectURL.Query().Get("state")
	fmt.Println("Code:", code, "State:", state)

	// Token exchange
	tokenReqBody := fmt.Sprintf("grant_type=authorization_code&code=%s", code)
	tokenResp, err := http.Post(
		baseURL+"/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(tokenReqBody),
	)
	if err != nil {
		log.Fatalf("Token request failed: %v", err)
	}
	defer tokenResp.Body.Close()

	tokenBody, _ := io.ReadAll(tokenResp.Body)
	fmt.Println("Token response:", string(tokenBody))

	//Access protected resource
	req, _ := http.NewRequest("GET", baseURL+"/hello", nil)
	req.Header.Set("Authorization", "Bearer mock-access-token-from-code")

	protectedResp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Protected request failed: %v", err)
	}
	defer protectedResp.Body.Close()

	protectedBody, _ := io.ReadAll(protectedResp.Body)
	fmt.Println("Protected resource response:", string(protectedBody))
}
