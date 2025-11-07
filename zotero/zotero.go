package zotero

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

// LibraryType represents the type of Zotero library (user or group)
type LibraryType string

const (
	LibraryTypeUser  LibraryType = "users"
	LibraryTypeGroup LibraryType = "groups"
)

// Client represents a Zotero API client
type Client struct {
	BaseURL      string
	LibraryID    string
	LibraryType  LibraryType
	APIKey       string
	Locale       string
	Timeout      time.Duration
	RateLimit    time.Duration
	RetryConfig  *RetryConfig
	httpClient   *http.Client
	rateLimiter  *rate.Limiter
	preserveJSON bool
	logger       *log.Logger
}

// RetryConfig defines retry behavior for failed requests
type RetryConfig struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	Jitter          bool
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// NewClient creates a new Zotero API client with the given library ID, library type, and options
func NewClient(libraryID string, libraryType LibraryType, opts ...ClientOption) *Client {
	client := &Client{
		BaseURL:      "https://api.zotero.org",
		LibraryID:    libraryID,
		LibraryType:  libraryType,
		Locale:       "en-US",
		Timeout:      30 * time.Second,
		RateLimit:    time.Second,
		httpClient:   &http.Client{},
		preserveJSON: false,
		logger:       log.New(io.Discard, "", 0),
	}

	for _, opt := range opts {
		opt(client)
	}

	// Configure HTTP client timeout
	client.httpClient.Timeout = client.Timeout

	// Configure rate limiter if rate limit is set
	if client.RateLimit > 0 {
		client.rateLimiter = rate.NewLimiter(rate.Every(client.RateLimit), 1)
	}

	return client
}

// WithAPIKey sets the API key for authentication
func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.APIKey = apiKey
	}
}

// WithBaseURL sets a custom base URL (e.g., for local Zotero server)
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// WithLocale sets the localization for the client
func WithLocale(locale string) ClientOption {
	return func(c *Client) {
		c.Locale = locale
	}
}

// WithTimeout sets the HTTP request timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.Timeout = timeout
	}
}

// WithRateLimit sets the rate limit for API requests
func WithRateLimit(rateLimit time.Duration) ClientOption {
	return func(c *Client) {
		c.RateLimit = rateLimit
	}
}

// WithRetry sets the retry configuration for failed requests
func WithRetry(config RetryConfig) ClientOption {
	return func(c *Client) {
		c.RetryConfig = &config
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithPreserveJSON sets whether to preserve JSON order
func WithPreserveJSON(preserve bool) ClientOption {
	return func(c *Client) {
		c.preserveJSON = preserve
	}
}

// WithLogger sets a custom logger for the client
func WithLogger(logger *log.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// joinWithOR joins string slices with OR operator (||)
func joinWithOR(values []string) string {
	if len(values) == 0 {
		return ""
	}
	result := values[0]
	for i := 1; i < len(values); i++ {
		result += " || " + values[i]
	}
	return result
}

// buildQueryString constructs URL query parameters
func (c *Client) buildQueryString(params *QueryParams) string {
	if params == nil {
		return ""
	}

	values := url.Values{}

	if params.Limit > 0 {
		values.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Start > 0 {
		values.Set("start", strconv.Itoa(params.Start))
	}
	if params.Sort != "" {
		values.Set("sort", params.Sort)
	}
	if params.Format != "" {
		values.Set("format", params.Format)
	}
	if params.Include != "" {
		values.Set("include", params.Include)
	}
	if params.Style != "" {
		values.Set("style", params.Style)
	}
	if params.Q != "" {
		values.Set("q", params.Q)
	}
	if params.QMode != "" {
		values.Set("qmode", params.QMode)
	}
	if params.Since > 0 {
		values.Set("since", strconv.Itoa(params.Since))
	}

	// Tags: Join multiple tags with OR operator (||)
	if len(params.Tag) > 0 {
		values.Set("tag", joinWithOR(params.Tag))
	}

	// ItemKeys: Join with comma separator (up to 50 items)
	if len(params.ItemKey) > 0 {
		itemKeyValue := params.ItemKey[0]
		for i := 1; i < len(params.ItemKey); i++ {
			itemKeyValue += "," + params.ItemKey[i]
		}
		values.Set("itemKey", itemKeyValue)
	}

	// ItemTypes: Join multiple item types with OR operator (||)
	if len(params.ItemType) > 0 {
		values.Set("itemType", joinWithOR(params.ItemType))
	}

	for k, v := range params.Extra {
		values.Set(k, v)
	}

	if query := values.Encode(); query != "" {
		return "?" + query
	}
	return ""
}

// doRequest performs an HTTP request with rate limiting and retries
func (c *Client) doRequest(ctx context.Context, method, path string, params *QueryParams) ([]byte, *http.Response, error) {
	// Apply rate limiting
	if c.rateLimiter != nil {
		c.logger.Printf("Waiting for rate limiter...")
		if err := c.rateLimiter.Wait(ctx); err != nil {
			c.logger.Printf("Rate limiter error: %v", err)
			return nil, nil, fmt.Errorf("rate limiter error: %w", err)
		}
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/%s/%s%s%s",
		c.BaseURL,
		c.LibraryType,
		c.LibraryID,
		path,
		c.buildQueryString(params),
	)

	c.logger.Printf("Making request: %s %s", method, urlStr)

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, urlStr, nil)
	if err != nil {
		c.logger.Printf("Error creating request: %v", err)
		return nil, nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	if c.APIKey != "" {
		req.Header.Set("Zotero-API-Key", c.APIKey)
		c.logger.Printf("API Key set: %s...", c.APIKey[:min(10, len(c.APIKey))])
	} else {
		c.logger.Printf("No API Key set")
	}
	req.Header.Set("Zotero-API-Version", "3")

	// Execute request
	c.logger.Printf("Executing request...")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Printf("Error executing request: %v", err)
		return nil, nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	c.logger.Printf("Response status: %d %s", resp.StatusCode, resp.Status)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Printf("Error reading response body: %v", err)
		return nil, resp, fmt.Errorf("error reading response body: %w", err)
	}

	c.logger.Printf("Response body length: %d bytes", len(body))

	// Check for errors
	if resp.StatusCode >= 400 {
		c.logger.Printf("API error: %s (status %d)", string(body), resp.StatusCode)
		return body, resp, fmt.Errorf("API error: %s (status %d)", string(body), resp.StatusCode)
	}

	c.logger.Printf("Request successful")
	return body, resp, nil
}
