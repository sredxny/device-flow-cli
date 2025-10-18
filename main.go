package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/sredxny/device-flow-cli/internal/auth"
)

var (
	clientID     string
	clientSecret string
	authURL      string
	tokenURL     string
	redirectURL  string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "device-flow-cli",
		Short: "A CLI tool with OAuth browser-based authentication",
		Long:  `A command-line tool that authenticates users via browser-based OAuth flow.`,
	}

	var loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Authenticate via browser-based OAuth flow",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runLogin(); err != nil {
				log.Fatalf("Login failed: %v", err)
			}
		},
	}

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Check authentication status",
		Run: func(cmd *cobra.Command, args []string) {
			if err := checkStatus(); err != nil {
				log.Fatalf("Status check failed: %v", err)
			}
		},
	}

	var logoutCmd = &cobra.Command{
		Use:   "logout",
		Short: "Clear stored authentication credentials",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runLogout(); err != nil {
				log.Fatalf("Logout failed: %v", err)
			}
		},
	}

	loginCmd.Flags().StringVar(&clientID, "client-id", os.Getenv("OAUTH_CLIENT_ID"), "OAuth client ID")
	loginCmd.Flags().StringVar(&clientSecret, "client-secret", os.Getenv("OAUTH_CLIENT_SECRET"), "OAuth client secret")
	loginCmd.Flags().StringVar(&authURL, "auth-url", os.Getenv("OAUTH_AUTH_URL"), "OAuth authorization URL")
	loginCmd.Flags().StringVar(&tokenURL, "token-url", os.Getenv("OAUTH_TOKEN_URL"), "OAuth token URL")
	loginCmd.Flags().StringVar(&redirectURL, "redirect-url", "http://localhost:8080/callback", "OAuth redirect URL")

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logoutCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runLogin() error {
	if clientID == "" || clientSecret == "" || authURL == "" || tokenURL == "" {
		return fmt.Errorf("missing required OAuth configuration. Set via flags or environment variables")
	}

	config := auth.OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      authURL,
		TokenURL:     tokenURL,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "profile", "email"},
	}

	authenticator := auth.NewAuthenticator(config)

	ctx := context.Background()
	token, err := authenticator.Login(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println("✓ Authentication successful!")
	fmt.Printf("Access token: %s...\n", token.AccessToken[:20])

	return nil
}

func checkStatus() error {
	token, err := auth.LoadToken()
	if err != nil {
		fmt.Println("Not authenticated. Run 'device-flow-cli login' to authenticate.")
		return nil
	}

	if token.Valid() {
		fmt.Println("✓ Authenticated")
		fmt.Printf("Token expires at: %s\n", token.Expiry.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("✗ Token expired. Run 'device-flow-cli login' to re-authenticate.")
	}

	return nil
}

func runLogout() error {
	if err := auth.ClearToken(); err != nil {
		return fmt.Errorf("failed to clear credentials: %w", err)
	}

	fmt.Println("✓ Logged out successfully")
	return nil
}
