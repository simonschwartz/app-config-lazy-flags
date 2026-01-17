package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonschwartz/app-config-lazy-flags/internal/appconfig"
	"github.com/simonschwartz/app-config-lazy-flags/internal/filecache"
)

func Run() {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	ctx := context.TODO()
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	cacheClient, err := filecache.New()
	if err != nil {
		log.Fatal(err)
	}
	client := appconfig.New(cfg)

	p := tea.NewProgram(
		NewModel(client, cacheClient),
		tea.WithAltScreen(),       // Use alternate screen buffer (full screen)
		// tea.WithMouseCellMotion(), // Enable mouse support
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Failed to run: %v", err)
		os.Exit(1)
	}
}

func isCredentialError(err error) bool {
	// Check for HTTP 403 Forbidden
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		// Common AWS authentication/authorization error codes
		switch code {
		case "ExpiredToken", "InvalidClientTokenId", "UnrecognizedClientException",
			"AccessDeniedException", "InvalidAccessKeyId", "SignatureDoesNotMatch":
			return true
		}

		// Check HTTP status code for 403
		var httpErr interface{ HTTPStatusCode() int }
		if errors.As(err, &httpErr) {
			return httpErr.HTTPStatusCode() == 403
		}
	}

	return false
}
