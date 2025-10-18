package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	RedirectURL  string
	Scopes       []string
}

// Authenticator handles OAuth authentication flow
type Authenticator struct {
	config      OAuthConfig
	oauth2Config *oauth2.Config
}

// NewAuthenticator creates a new Authenticator instance
func NewAuthenticator(config OAuthConfig) *Authenticator {
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL,
			TokenURL: config.TokenURL,
		},
	}

	return &Authenticator{
		config:      config,
		oauth2Config: oauth2Config,
	}
}

// Login initiates the OAuth flow and returns a token
func (a *Authenticator) Login(ctx context.Context) (*oauth2.Token, error) {
	// Generate random state for CSRF protection
	state, err := generateRandomState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Start local callback server
	tokenChan := make(chan *oauth2.Token, 1)
	errChan := make(chan error, 1)

	server := a.startCallbackServer(state, tokenChan, errChan)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	// Generate authorization URL
	authURL := a.oauth2Config.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)

	fmt.Println("Opening browser for authentication...")
	fmt.Printf("If browser doesn't open automatically, visit:\n%s\n\n", authURL)

	// Open browser
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Failed to open browser automatically: %v\n", err)
	}

	// Wait for callback or error
	select {
	case token := <-tokenChan:
		// Save token
		if err := SaveToken(token); err != nil {
			return nil, fmt.Errorf("failed to save token: %w", err)
		}
		return token, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authentication timeout")
	}
}

// startCallbackServer starts a local HTTP server to handle the OAuth callback
func (a *Authenticator) startCallbackServer(expectedState string, tokenChan chan *oauth2.Token, errChan chan error) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		errorParam := r.URL.Query().Get("error")

		// Handle OAuth errors
		if errorParam != "" {
			errorDesc := r.URL.Query().Get("error_description")
			err := fmt.Errorf("oauth error: %s - %s", errorParam, errorDesc)
			errChan <- err
			http.Error(w, "Authentication failed. You can close this window.", http.StatusBadRequest)
			return
		}

		// Verify state
		if state != expectedState {
			errChan <- fmt.Errorf("invalid state parameter")
			http.Error(w, "Authentication failed: invalid state. You can close this window.", http.StatusBadRequest)
			return
		}

		// Exchange code for token
		ctx := context.Background()
		token, err := a.oauth2Config.Exchange(ctx, code)
		if err != nil {
			errChan <- fmt.Errorf("failed to exchange code for token: %w", err)
			http.Error(w, "Authentication failed. You can close this window.", http.StatusInternalServerError)
			return
		}

		// Send success response
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<html>
				<head><title>Authentication Successful</title></head>
				<body>
					<h1>✓ Authentication Successful!</h1>
					<p>You can close this window and return to the CLI.</p>
					<script>window.close();</script>
				</body>
			</html>
		`)

		tokenChan <- token
	})

	// Parse redirect URL to get port
	parsedURL, _ := url.Parse(a.config.RedirectURL)
	addr := parsedURL.Host
	if addr == "" {
		addr = "localhost:8080"
	}

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	return server
}

// generateRandomState generates a random state string for CSRF protection
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
