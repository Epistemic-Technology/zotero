package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/Epistemic-Technology/zotero/zotero"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	APIKey      string
	LibraryID   string
	LibraryType zotero.LibraryType
	BaseURL     string
}

// getTestConfig loads test configuration from environment variables.
// Returns nil if required credentials are not set.
func getTestConfig() *TestConfig {
	apiKey := os.Getenv("ZOTERO_API_KEY")
	libraryID := os.Getenv("ZOTERO_LIBRARY_ID")
	libraryType := os.Getenv("ZOTERO_LIBRARY_TYPE")
	baseURL := os.Getenv("TEST_API_URL")

	// Required fields
	if apiKey == "" || libraryID == "" {
		return nil
	}

	// Default to user library type
	if libraryType == "" {
		libraryType = "user"
	}

	// Default to live Zotero API
	if baseURL == "" {
		baseURL = "https://api.zotero.org"
	}

	var libType zotero.LibraryType
	switch strings.ToLower(libraryType) {
	case "user":
		libType = zotero.LibraryTypeUser
	case "group":
		libType = zotero.LibraryTypeGroup
	default:
		libType = zotero.LibraryTypeUser
	}

	return &TestConfig{
		APIKey:      apiKey,
		LibraryID:   libraryID,
		LibraryType: libType,
		BaseURL:     baseURL,
	}
}

// newTestClient creates a new Zotero client configured for integration testing.
// Returns nil if credentials are not available.
func newTestClient() *zotero.Client {
	config := getTestConfig()
	if config == nil {
		return nil
	}

	return zotero.NewClient(
		config.LibraryID,
		config.LibraryType,
		zotero.WithAPIKey(config.APIKey),
		zotero.WithBaseURL(config.BaseURL),
		zotero.WithRateLimit(0), // Disable rate limiting for faster tests
	)
}

// skipIfNoCredentials skips the test if integration test credentials are not available
func skipIfNoCredentials(t *testing.T) *zotero.Client {
	t.Helper()

	client := newTestClient()
	if client == nil {
		t.Skip("Skipping integration test: ZOTERO_API_KEY and ZOTERO_LIBRARY_ID not set")
	}

	return client
}

// isLocalAPI returns true if testing against a local REST API
func isLocalAPI() bool {
	baseURL := os.Getenv("TEST_API_URL")
	return strings.Contains(baseURL, "localhost") || strings.Contains(baseURL, "127.0.0.1")
}

// getTestLibraryType returns the configured library type for tests
func getTestLibraryType() string {
	libraryType := os.Getenv("ZOTERO_LIBRARY_TYPE")
	if libraryType == "" {
		return "user"
	}
	return strings.ToLower(libraryType)
}
