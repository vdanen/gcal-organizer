// Package auth provides authentication for Google Workspace APIs and Gemini.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/tasks/v1"
)

// Scopes required for Google Workspace APIs
var Scopes = []string{
	drive.DriveScope,
	docs.DocumentsScope,
	calendar.CalendarReadonlyScope,
	tasks.TasksScope,
}

// OAuthClient handles OAuth2 authentication for Google Workspace APIs.
type OAuthClient struct {
	config     *oauth2.Config
	tokenFile  string
	httpClient *http.Client
}

// NewOAuthClient creates a new OAuth client from credentials file.
func NewOAuthClient(credentialsFile, tokenFile string) (*OAuthClient, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w", err)
	}

	config, err := google.ConfigFromJSON(b, Scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %w", err)
	}

	return &OAuthClient{
		config:    config,
		tokenFile: tokenFile,
	}, nil
}

// GetClient returns an authenticated HTTP client.
// If no cached token exists, it will prompt for authorization.
func (o *OAuthClient) GetClient(ctx context.Context) (*http.Client, error) {
	if o.httpClient != nil {
		return o.httpClient, nil
	}

	tok, err := o.loadToken()
	if err != nil {
		// No saved token, need to get one
		tok, err = o.getTokenFromWeb(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get token: %w", err)
		}
		if err := o.saveToken(tok); err != nil {
			return nil, fmt.Errorf("unable to save token: %w", err)
		}
	}

	o.httpClient = o.config.Client(ctx, tok)
	return o.httpClient, nil
}

// loadToken reads the cached OAuth token from file.
func (o *OAuthClient) loadToken() (*oauth2.Token, error) {
	f, err := os.Open(o.tokenFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken saves the OAuth token to file.
func (o *OAuthClient) saveToken(token *oauth2.Token) error {
	// Ensure directory exists
	dir := o.tokenFile[:len(o.tokenFile)-len("/token.json")]
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("unable to create token directory: %w", err)
	}

	f, err := os.OpenFile(o.tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to create token file: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

// getTokenFromWeb starts an OAuth2 flow in the browser.
func (o *OAuthClient) getTokenFromWeb(ctx context.Context) (*oauth2.Token, error) {
	authURL := o.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Print("🔗 Follow these steps to authorize gcal-organizer:\n\n")
	fmt.Print("  1. Open this URL in your browser:\n\n")
	fmt.Printf("     %v\n\n", authURL)
	fmt.Println("  2. Sign in with your Google account and click 'Allow'")
	fmt.Println("  3. You'll see a page saying \"This site can't be reached\"")
	fmt.Println("     — that's expected!")
	fmt.Println("  4. Look at the URL in your browser's address bar.")
	fmt.Println("     Find the part after 'code=' and before '&scope='")
	fmt.Print("     Copy that entire code.\n\n")
	fmt.Println("     Example URL: http://localhost/?code=4/0AXSc3g...abc&scope=...")
	fmt.Print("     The code is:                         4/0AXSc3g...abc\n\n")
	fmt.Print("📝 Paste the authorization code here: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	tok, err := o.config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to exchange code for token: %w", err)
	}
	return tok, nil
}

// IsAuthenticated checks if a valid token exists.
func (o *OAuthClient) IsAuthenticated() bool {
	tok, err := o.loadToken()
	if err != nil {
		return false
	}
	return tok.Valid()
}
