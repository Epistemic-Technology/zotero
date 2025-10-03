package zotero

import (
	"log"
	"net/http"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		libraryID   string
		libraryType LibraryType
		opts        []ClientOption
		wantBaseURL string
		wantLocale  string
		wantTimeout time.Duration
	}{
		{
			name:        "default client",
			libraryID:   "12345",
			libraryType: LibraryTypeUser,
			opts:        nil,
			wantBaseURL: "https://api.zotero.org",
			wantLocale:  "en-US",
			wantTimeout: 30 * time.Second,
		},
		{
			name:        "with API key",
			libraryID:   "12345",
			libraryType: LibraryTypeUser,
			opts:        []ClientOption{WithAPIKey("test-key")},
			wantBaseURL: "https://api.zotero.org",
			wantLocale:  "en-US",
			wantTimeout: 30 * time.Second,
		},
		{
			name:        "with custom base URL",
			libraryID:   "12345",
			libraryType: LibraryTypeUser,
			opts:        []ClientOption{WithBaseURL("https://custom.example.com")},
			wantBaseURL: "https://custom.example.com",
			wantLocale:  "en-US",
			wantTimeout: 30 * time.Second,
		},
		{
			name:        "with custom locale",
			libraryID:   "12345",
			libraryType: LibraryTypeUser,
			opts:        []ClientOption{WithLocale("de-DE")},
			wantBaseURL: "https://api.zotero.org",
			wantLocale:  "de-DE",
			wantTimeout: 30 * time.Second,
		},
		{
			name:        "with custom timeout",
			libraryID:   "12345",
			libraryType: LibraryTypeUser,
			opts:        []ClientOption{WithTimeout(60 * time.Second)},
			wantBaseURL: "https://api.zotero.org",
			wantLocale:  "en-US",
			wantTimeout: 60 * time.Second,
		},
		{
			name:        "group library type",
			libraryID:   "67890",
			libraryType: LibraryTypeGroup,
			opts:        nil,
			wantBaseURL: "https://api.zotero.org",
			wantLocale:  "en-US",
			wantTimeout: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.libraryID, tt.libraryType, tt.opts...)

			if client.LibraryID != tt.libraryID {
				t.Errorf("LibraryID = %v, want %v", client.LibraryID, tt.libraryID)
			}
			if client.LibraryType != tt.libraryType {
				t.Errorf("LibraryType = %v, want %v", client.LibraryType, tt.libraryType)
			}
			if client.BaseURL != tt.wantBaseURL {
				t.Errorf("BaseURL = %v, want %v", client.BaseURL, tt.wantBaseURL)
			}
			if client.Locale != tt.wantLocale {
				t.Errorf("Locale = %v, want %v", client.Locale, tt.wantLocale)
			}
			if client.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %v, want %v", client.Timeout, tt.wantTimeout)
			}
			if client.httpClient == nil {
				t.Error("httpClient should not be nil")
			}
			if client.logger == nil {
				t.Error("logger should not be nil")
			}
		})
	}
}

func TestWithAPIKey(t *testing.T) {
	apiKey := "test-api-key-123"
	client := NewClient("12345", LibraryTypeUser, WithAPIKey(apiKey))

	if client.APIKey != apiKey {
		t.Errorf("APIKey = %v, want %v", client.APIKey, apiKey)
	}
}

func TestWithRateLimit(t *testing.T) {
	rateLimit := 2 * time.Second
	client := NewClient("12345", LibraryTypeUser, WithRateLimit(rateLimit))

	if client.RateLimit != rateLimit {
		t.Errorf("RateLimit = %v, want %v", client.RateLimit, rateLimit)
	}
	if client.rateLimiter == nil {
		t.Error("rateLimiter should not be nil when RateLimit is set")
	}
}

func TestWithRetry(t *testing.T) {
	retryConfig := RetryConfig{
		MaxAttempts:     3,
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     5 * time.Second,
		Multiplier:      2.0,
		Jitter:          true,
	}
	client := NewClient("12345", LibraryTypeUser, WithRetry(retryConfig))

	if client.RetryConfig == nil {
		t.Fatal("RetryConfig should not be nil")
	}
	if client.RetryConfig.MaxAttempts != retryConfig.MaxAttempts {
		t.Errorf("RetryConfig.MaxAttempts = %v, want %v", client.RetryConfig.MaxAttempts, retryConfig.MaxAttempts)
	}
	if client.RetryConfig.InitialInterval != retryConfig.InitialInterval {
		t.Errorf("RetryConfig.InitialInterval = %v, want %v", client.RetryConfig.InitialInterval, retryConfig.InitialInterval)
	}
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	client := NewClient("12345", LibraryTypeUser, WithHTTPClient(customClient))

	if client.httpClient != customClient {
		t.Error("httpClient should be the custom client")
	}
}

func TestWithPreserveJSON(t *testing.T) {
	client := NewClient("12345", LibraryTypeUser, WithPreserveJSON(true))

	if !client.preserveJSON {
		t.Error("preserveJSON should be true")
	}
}

func TestWithLogger(t *testing.T) {
	customLogger := log.New(log.Writer(), "TEST: ", log.LstdFlags)
	client := NewClient("12345", LibraryTypeUser, WithLogger(customLogger))

	if client.logger != customLogger {
		t.Error("logger should be the custom logger")
	}
}

func TestBuildQueryString(t *testing.T) {
	client := NewClient("12345", LibraryTypeUser)

	tests := []struct {
		name   string
		params *QueryParams
		want   string
	}{
		{
			name:   "nil params",
			params: nil,
			want:   "",
		},
		{
			name:   "empty params",
			params: &QueryParams{},
			want:   "",
		},
		{
			name: "with limit",
			params: &QueryParams{
				Limit: 10,
			},
			want: "?limit=10",
		},
		{
			name: "with limit and start",
			params: &QueryParams{
				Limit: 10,
				Start: 20,
			},
			want: "?limit=10&start=20",
		},
		{
			name: "with sort",
			params: &QueryParams{
				Sort: "title",
			},
			want: "?sort=title",
		},
		{
			name: "with format",
			params: &QueryParams{
				Format: "json",
			},
			want: "?format=json",
		},
		{
			name: "with tags",
			params: &QueryParams{
				Tag: []string{"tag1", "tag2"},
			},
			want: "?tag=tag1&tag=tag2",
		},
		{
			name: "with item keys",
			params: &QueryParams{
				ItemKey: []string{"KEY1", "KEY2"},
			},
			want: "?itemKey=KEY1&itemKey=KEY2",
		},
		{
			name: "with since",
			params: &QueryParams{
				Since: 100,
			},
			want: "?since=100",
		},
		{
			name: "with extra params",
			params: &QueryParams{
				Extra: map[string]string{
					"custom": "value",
				},
			},
			want: "?custom=value",
		},
		{
			name: "with multiple params",
			params: &QueryParams{
				Limit:   50,
				Start:   100,
				Sort:    "dateModified",
				Format:  "atom",
				Include: "data",
				Style:   "apa",
				Q:       "search query",
				QMode:   "everything",
			},
			want: "?format=atom&include=data&limit=50&q=search+query&qmode=everything&sort=dateModified&start=100&style=apa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.buildQueryString(tt.params)
			if got != tt.want {
				t.Errorf("buildQueryString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLibraryType(t *testing.T) {
	if LibraryTypeUser != "users" {
		t.Errorf("LibraryTypeUser = %v, want 'users'", LibraryTypeUser)
	}
	if LibraryTypeGroup != "groups" {
		t.Errorf("LibraryTypeGroup = %v, want 'groups'", LibraryTypeGroup)
	}
}
