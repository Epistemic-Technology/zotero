package zotero

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// CreateItems creates one or more items in the library.
// Accepts up to 50 items per request.
// Returns the write response indicating success, unchanged, and failed items.
func (c *Client) CreateItems(ctx context.Context, items []Item) (*WriteResponse, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no items provided")
	}
	if len(items) > 50 {
		return nil, fmt.Errorf("maximum 50 items per request, got %d", len(items))
	}

	// Extract just the data portion for creation
	itemsData := make([]ItemData, len(items))
	for i, item := range items {
		itemsData[i] = item.Data
	}

	body, err := json.Marshal(itemsData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling items: %w", err)
	}

	respBody, resp, err := c.doWriteRequest(ctx, http.MethodPost, "/items", body, 0)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var writeResp WriteResponse
	if err := json.Unmarshal(respBody, &writeResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &writeResp, nil
}

// UpdateItem updates a single item in the library.
// The item must contain version information for concurrency control.
// Returns nil on success, error otherwise.
func (c *Client) UpdateItem(ctx context.Context, item *Item) error {
	if item == nil {
		return fmt.Errorf("item cannot be nil")
	}
	if item.Key == "" && item.Data.Key == "" {
		return fmt.Errorf("item key is required")
	}

	key := item.Key
	if key == "" {
		key = item.Data.Key
	}

	version := item.Version
	if version == 0 {
		version = item.Data.Version
	}

	body, err := json.Marshal(item.Data)
	if err != nil {
		return fmt.Errorf("error marshaling item: %w", err)
	}

	path := fmt.Sprintf("/items/%s", key)
	respBody, resp, err := c.doWriteRequest(ctx, http.MethodPatch, path, body, version)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// UpdateItems updates multiple items in the library (up to 50 items).
// Each item must contain version information for concurrency control.
// Returns the write response indicating success, unchanged, and failed items.
func (c *Client) UpdateItems(ctx context.Context, items []Item) (*WriteResponse, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no items provided")
	}
	if len(items) > 50 {
		return nil, fmt.Errorf("maximum 50 items per request, got %d", len(items))
	}

	// For batch updates, we need to include the key and version
	itemsData := make([]map[string]any, len(items))
	for i, item := range items {
		key := item.Key
		if key == "" {
			key = item.Data.Key
		}
		version := item.Version
		if version == 0 {
			version = item.Data.Version
		}

		if key == "" {
			return nil, fmt.Errorf("item %d missing key", i)
		}
		if version == 0 {
			return nil, fmt.Errorf("item %d missing version", i)
		}

		// Marshal to map to include key and version
		data := make(map[string]any)
		dataBytes, err := json.Marshal(item.Data)
		if err != nil {
			return nil, fmt.Errorf("error marshaling item %d: %w", i, err)
		}
		if err := json.Unmarshal(dataBytes, &data); err != nil {
			return nil, fmt.Errorf("error unmarshaling item %d: %w", i, err)
		}

		data["key"] = key
		data["version"] = version
		itemsData[i] = data
	}

	body, err := json.Marshal(itemsData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling items: %w", err)
	}

	respBody, resp, err := c.doWriteRequest(ctx, http.MethodPost, "/items", body, 0)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var writeResp WriteResponse
	if err := json.Unmarshal(respBody, &writeResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &writeResp, nil
}

// DeleteItem deletes a single item from the library.
// The item must contain version information for concurrency control.
// Returns nil on success, error otherwise.
func (c *Client) DeleteItem(ctx context.Context, itemKey string, version int) error {
	if itemKey == "" {
		return fmt.Errorf("item key is required")
	}
	if version == 0 {
		return fmt.Errorf("version is required for delete operations")
	}

	path := fmt.Sprintf("/items/%s", itemKey)
	respBody, resp, err := c.doWriteRequest(ctx, http.MethodDelete, path, nil, version)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DeleteItems deletes multiple items from the library (up to 50 items).
// Each item key must have a corresponding version for concurrency control.
// Returns nil on success, error otherwise.
func (c *Client) DeleteItems(ctx context.Context, itemKeys []string, version int) error {
	if len(itemKeys) == 0 {
		return fmt.Errorf("no item keys provided")
	}
	if len(itemKeys) > 50 {
		return fmt.Errorf("maximum 50 items per request, got %d", len(itemKeys))
	}
	if version == 0 {
		return fmt.Errorf("version is required for delete operations")
	}

	// Multiple deletes use itemKey query parameter
	path := fmt.Sprintf("/items?itemKey=%s", strings.Join(itemKeys, ","))
	respBody, resp, err := c.doWriteRequest(ctx, http.MethodDelete, path, nil, version)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// CreateCollections creates one or more collections in the library.
// Accepts up to 50 collections per request.
// Returns the write response indicating success, unchanged, and failed collections.
func (c *Client) CreateCollections(ctx context.Context, collections []Collection) (*WriteResponse, error) {
	if len(collections) == 0 {
		return nil, fmt.Errorf("no collections provided")
	}
	if len(collections) > 50 {
		return nil, fmt.Errorf("maximum 50 collections per request, got %d", len(collections))
	}

	// Extract just the data portion for creation
	collectionsData := make([]CollectionData, len(collections))
	for i, coll := range collections {
		collectionsData[i] = coll.Data
	}

	body, err := json.Marshal(collectionsData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling collections: %w", err)
	}

	respBody, resp, err := c.doWriteRequest(ctx, http.MethodPost, "/collections", body, 0)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var writeResp WriteResponse
	if err := json.Unmarshal(respBody, &writeResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &writeResp, nil
}

// UpdateCollection updates a single collection in the library.
// The collection must contain version information for concurrency control.
// Returns nil on success, error otherwise.
func (c *Client) UpdateCollection(ctx context.Context, collection *Collection) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}
	if collection.Key == "" && collection.Data.Key == "" {
		return fmt.Errorf("collection key is required")
	}

	key := collection.Key
	if key == "" {
		key = collection.Data.Key
	}

	version := collection.Version
	if version == 0 {
		version = collection.Data.Version
	}

	body, err := json.Marshal(collection.Data)
	if err != nil {
		return fmt.Errorf("error marshaling collection: %w", err)
	}

	path := fmt.Sprintf("/collections/%s", key)
	respBody, resp, err := c.doWriteRequest(ctx, http.MethodPatch, path, body, version)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// UpdateCollections updates multiple collections in the library (up to 50 collections).
// Each collection must contain version information for concurrency control.
// Returns the write response indicating success, unchanged, and failed collections.
func (c *Client) UpdateCollections(ctx context.Context, collections []Collection) (*WriteResponse, error) {
	if len(collections) == 0 {
		return nil, fmt.Errorf("no collections provided")
	}
	if len(collections) > 50 {
		return nil, fmt.Errorf("maximum 50 collections per request, got %d", len(collections))
	}

	// For batch updates, we need to include the key and version
	collectionsData := make([]map[string]any, len(collections))
	for i, coll := range collections {
		key := coll.Key
		if key == "" {
			key = coll.Data.Key
		}
		version := coll.Version
		if version == 0 {
			version = coll.Data.Version
		}

		if key == "" {
			return nil, fmt.Errorf("collection %d missing key", i)
		}
		if version == 0 {
			return nil, fmt.Errorf("collection %d missing version", i)
		}

		// Marshal to map to include key and version
		data := make(map[string]any)
		dataBytes, err := json.Marshal(coll.Data)
		if err != nil {
			return nil, fmt.Errorf("error marshaling collection %d: %w", i, err)
		}
		if err := json.Unmarshal(dataBytes, &data); err != nil {
			return nil, fmt.Errorf("error unmarshaling collection %d: %w", i, err)
		}

		data["key"] = key
		data["version"] = version
		collectionsData[i] = data
	}

	body, err := json.Marshal(collectionsData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling collections: %w", err)
	}

	respBody, resp, err := c.doWriteRequest(ctx, http.MethodPost, "/collections", body, 0)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var writeResp WriteResponse
	if err := json.Unmarshal(respBody, &writeResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &writeResp, nil
}

// DeleteCollection deletes a single collection from the library.
// The collection must exist and version must match for concurrency control.
// Returns nil on success, error otherwise.
func (c *Client) DeleteCollection(ctx context.Context, collectionKey string, version int) error {
	if collectionKey == "" {
		return fmt.Errorf("collection key is required")
	}
	if version == 0 {
		return fmt.Errorf("version is required for delete operations")
	}

	path := fmt.Sprintf("/collections/%s", collectionKey)
	respBody, resp, err := c.doWriteRequest(ctx, http.MethodDelete, path, nil, version)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DeleteCollections deletes multiple collections from the library (up to 50 collections).
// Each collection key must have a corresponding version for concurrency control.
// Returns nil on success, error otherwise.
func (c *Client) DeleteCollections(ctx context.Context, collectionKeys []string, version int) error {
	if len(collectionKeys) == 0 {
		return fmt.Errorf("no collection keys provided")
	}
	if len(collectionKeys) > 50 {
		return fmt.Errorf("maximum 50 collections per request, got %d", len(collectionKeys))
	}
	if version == 0 {
		return fmt.Errorf("version is required for delete operations")
	}

	// Multiple deletes use collectionKey query parameter
	path := fmt.Sprintf("/collections?collectionKey=%s", strings.Join(collectionKeys, ","))
	respBody, resp, err := c.doWriteRequest(ctx, http.MethodDelete, path, nil, version)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// CreateSearches creates one or more saved searches in the library.
// Accepts up to 50 searches per request.
// Returns the write response indicating success, unchanged, and failed searches.
func (c *Client) CreateSearches(ctx context.Context, searches []Search) (*WriteResponse, error) {
	if len(searches) == 0 {
		return nil, fmt.Errorf("no searches provided")
	}
	if len(searches) > 50 {
		return nil, fmt.Errorf("maximum 50 searches per request, got %d", len(searches))
	}

	// Extract just the data portion for creation
	searchesData := make([]SearchData, len(searches))
	for i, search := range searches {
		searchesData[i] = search.Data
	}

	body, err := json.Marshal(searchesData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling searches: %w", err)
	}

	respBody, resp, err := c.doWriteRequest(ctx, http.MethodPost, "/searches", body, 0)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var writeResp WriteResponse
	if err := json.Unmarshal(respBody, &writeResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &writeResp, nil
}

// UpdateSearch updates a single saved search in the library.
// The search must contain version information for concurrency control.
// Returns nil on success, error otherwise.
func (c *Client) UpdateSearch(ctx context.Context, search *Search) error {
	if search == nil {
		return fmt.Errorf("search cannot be nil")
	}
	if search.Key == "" && search.Data.Key == "" {
		return fmt.Errorf("search key is required")
	}

	key := search.Key
	if key == "" {
		key = search.Data.Key
	}

	version := search.Version
	if version == 0 {
		version = search.Data.Version
	}

	body, err := json.Marshal(search.Data)
	if err != nil {
		return fmt.Errorf("error marshaling search: %w", err)
	}

	path := fmt.Sprintf("/searches/%s", key)
	respBody, resp, err := c.doWriteRequest(ctx, http.MethodPatch, path, body, version)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DeleteSearch deletes a single saved search from the library.
// The search must exist and version must match for concurrency control.
// Returns nil on success, error otherwise.
func (c *Client) DeleteSearch(ctx context.Context, searchKey string, version int) error {
	if searchKey == "" {
		return fmt.Errorf("search key is required")
	}
	if version == 0 {
		return fmt.Errorf("version is required for delete operations")
	}

	path := fmt.Sprintf("/searches/%s", searchKey)
	respBody, resp, err := c.doWriteRequest(ctx, http.MethodDelete, path, nil, version)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DeleteSearches deletes multiple saved searches from the library (up to 50 searches).
// Each search key must have a corresponding version for concurrency control.
// Returns nil on success, error otherwise.
func (c *Client) DeleteSearches(ctx context.Context, searchKeys []string, version int) error {
	if len(searchKeys) == 0 {
		return fmt.Errorf("no search keys provided")
	}
	if len(searchKeys) > 50 {
		return fmt.Errorf("maximum 50 searches per request, got %d", len(searchKeys))
	}
	if version == 0 {
		return fmt.Errorf("version is required for delete operations")
	}

	// Multiple deletes use searchKey query parameter
	path := fmt.Sprintf("/searches?searchKey=%s", strings.Join(searchKeys, ","))
	respBody, resp, err := c.doWriteRequest(ctx, http.MethodDelete, path, nil, version)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// AddTags adds one or more tags to an item.
// This is a convenience method that fetches the item, adds the tags, and updates it.
// Returns nil on success, error otherwise.
func (c *Client) AddTags(ctx context.Context, itemKey string, tags ...string) error {
	if itemKey == "" {
		return fmt.Errorf("item key is required")
	}
	if len(tags) == 0 {
		return fmt.Errorf("no tags provided")
	}

	// Fetch the current item
	item, err := c.Item(ctx, itemKey, nil)
	if err != nil {
		return fmt.Errorf("error fetching item: %w", err)
	}

	// Add new tags (avoiding duplicates)
	existingTags := make(map[string]bool)
	for _, tag := range item.Data.Tags {
		existingTags[tag.Tag] = true
	}

	for _, tagName := range tags {
		if !existingTags[tagName] {
			item.Data.Tags = append(item.Data.Tags, Tag{Tag: tagName})
		}
	}

	// Update the item
	return c.UpdateItem(ctx, item)
}

// DeleteTags deletes tags from the library by name.
// This removes the tags from all items in the library.
// Returns nil on success, error otherwise.
func (c *Client) DeleteTags(ctx context.Context, version int, tags ...string) error {
	if len(tags) == 0 {
		return fmt.Errorf("no tags provided")
	}
	if version == 0 {
		return fmt.Errorf("version is required for delete operations")
	}

	// Multiple tag deletes use tag query parameter
	// URL encode the tags
	encodedTags := make([]string, len(tags))
	copy(tags, encodedTags)

	path := fmt.Sprintf("/tags?tag=%s", strings.Join(encodedTags, "&tag="))
	respBody, resp, err := c.doWriteRequest(ctx, http.MethodDelete, path, nil, version)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// UploadAttachment uploads a file as an attachment to a parent item.
// This is a multi-step process:
// 1. Create an attachment item with linkMode "imported_file" or "imported_url"
// 2. Get upload authorization
// 3. Upload the file
// 4. Register the upload
//
// parentItemKey: The key of the parent item to attach to (empty string for standalone attachment)
// filepath: Path to the file to upload
// filename: Name to use for the attachment (if empty, uses basename of filepath)
// contentType: MIME type of the file (e.g., "application/pdf")
func (c *Client) UploadAttachment(ctx context.Context, parentItemKey, filepath, filename, contentType string) (*Item, error) {
	// Read file for MD5 and size
	fileData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if filename == "" {
		filename = filepath[strings.LastIndex(filepath, "/")+1:]
	}

	// Calculate MD5
	md5Hash := md5.Sum(fileData)
	md5String := hex.EncodeToString(md5Hash[:])

	// Step 1: Create attachment item
	attachment := Item{
		Data: ItemData{
			ItemType:    ItemTypeAttachment,
			LinkMode:    "imported_file",
			Title:       filename,
			ContentType: contentType,
			Filename:    filename,
			MD5:         md5String,
			MTime:       time.Now().UnixMilli(),
		},
	}

	if parentItemKey != "" {
		attachment.Data.ParentItem = parentItemKey
	}

	resp, err := c.CreateItems(ctx, []Item{attachment})
	if err != nil {
		return nil, fmt.Errorf("error creating attachment item: %w", err)
	}

	if len(resp.Success) == 0 {
		if len(resp.Failed) > 0 {
			return nil, fmt.Errorf("failed to create attachment: %s", resp.Failed["0"].Message)
		}
		return nil, fmt.Errorf("failed to create attachment: no success or error reported")
	}

	// Get the attachment key from the response
	var attachmentKey string
	for _, keyVal := range resp.Success {
		if key, ok := keyVal.(string); ok {
			attachmentKey = key
			break
		}
	}

	// Step 2: Request upload authorization
	// Build form-encoded request body (not JSON!)
	authBody := []byte(fmt.Sprintf("md5=%s&filename=%s&filesize=%d&mtime=%d",
		md5String, filename, len(fileData), attachment.Data.MTime))

	path := fmt.Sprintf("/items/%s/file", attachmentKey)
	authRespBody, authResp, err := c.doFileAuthRequest(ctx, path, authBody, "*", "")

	// If we get a 412 with "file exists", try again with If-Match header using the file's MD5
	if err != nil && authResp != nil && authResp.StatusCode == http.StatusPreconditionFailed {
		c.logger.Printf("File exists on server (412), retrying with If-Match header")
		authRespBody, authResp, err = c.doFileAuthRequest(ctx, path, authBody, "", md5String)
	}

	if err != nil {
		return nil, fmt.Errorf("error requesting upload authorization: %w", err)
	}

	// Parse authorization response
	var authResponse map[string]any
	if err := json.Unmarshal(authRespBody, &authResponse); err != nil {
		return nil, fmt.Errorf("error parsing auth response: %w", err)
	}

	// Check if file already exists
	if exists, ok := authResponse["exists"].(float64); ok && exists == 1 {
		c.logger.Printf("File already exists on server")
		// Fetch and return the attachment item
		return c.Item(ctx, attachmentKey, nil)
	}

	// Step 3: Upload the file
	uploadURL, ok := authResponse["url"].(string)
	if !ok {
		return nil, fmt.Errorf("missing upload URL in auth response")
	}

	uploadParams, ok := authResponse["params"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing upload params in auth response")
	}

	// Create multipart form
	var uploadBody bytes.Buffer
	writer := multipart.NewWriter(&uploadBody)

	// Add form fields from params
	for key, val := range uploadParams {
		if valStr, ok := val.(string); ok {
			if err := writer.WriteField(key, valStr); err != nil {
				return nil, fmt.Errorf("error writing field %s: %w", key, err)
			}
		}
	}

	// Add the file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("error creating form file: %w", err)
	}
	if _, err := part.Write(fileData); err != nil {
		return nil, fmt.Errorf("error writing file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("error closing multipart writer: %w", err)
	}

	// Upload to S3/storage
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, &uploadBody)
	if err != nil {
		return nil, fmt.Errorf("error creating upload request: %w", err)
	}
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())

	uploadResp, err := c.httpClient.Do(uploadReq)
	if err != nil {
		return nil, fmt.Errorf("error uploading file: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK && uploadResp.StatusCode != http.StatusCreated && uploadResp.StatusCode != http.StatusNoContent {
		uploadRespBody, _ := io.ReadAll(uploadResp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", uploadResp.StatusCode, string(uploadRespBody))
	}

	// Step 4: Register the upload
	registerPath := fmt.Sprintf("/items/%s/file", attachmentKey)
	registerBody := []byte(fmt.Sprintf(`{"upload": "%s"}`, authResponse["uploadKey"]))

	if lastModified := authResp.Header.Get("Last-Modified-Version"); lastModified != "" {
		if version, err := strconv.Atoi(lastModified); err == nil {
			_, registerResp, err := c.doWriteRequest(ctx, http.MethodPost, registerPath, registerBody, version)
			if err != nil {
				return nil, fmt.Errorf("error registering upload: %w", err)
			}
			if registerResp.StatusCode != http.StatusNoContent {
				return nil, fmt.Errorf("unexpected status code from register: %d", registerResp.StatusCode)
			}
		}
	}

	// Fetch and return the final attachment item
	return c.Item(ctx, attachmentKey, nil)
}

// doFileAuthRequest performs an HTTP request to authorize file upload with If-Match/If-None-Match headers
func (c *Client) doFileAuthRequest(ctx context.Context, path string, body []byte, ifNoneMatch, ifMatch string) ([]byte, *http.Response, error) {
	// Apply rate limiting
	if c.rateLimiter != nil {
		c.logger.Printf("Waiting for rate limiter...")
		if err := c.rateLimiter.Wait(ctx); err != nil {
			c.logger.Printf("Rate limiter error: %v", err)
			return nil, nil, fmt.Errorf("rate limiter error: %w", err)
		}
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/%s/%s%s",
		c.BaseURL,
		c.LibraryType,
		c.LibraryID,
		path,
	)

	c.logger.Printf("Making file auth request: POST %s", urlStr)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(body))
	if err != nil {
		c.logger.Printf("Error creating request: %v", err)
		return nil, nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	if c.APIKey != "" {
		req.Header.Set("Zotero-API-Key", c.APIKey)
		c.logger.Printf("API Key set")
	}
	req.Header.Set("Zotero-API-Version", "3")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set If-Match or If-None-Match headers (required for file upload authorization)
	if ifNoneMatch != "" {
		req.Header.Set("If-None-Match", ifNoneMatch)
		c.logger.Printf("If-None-Match: %s", ifNoneMatch)
	} else if ifMatch != "" {
		req.Header.Set("If-Match", ifMatch)
		c.logger.Printf("If-Match: %s", ifMatch)
	}

	// Execute request
	c.logger.Printf("Executing file auth request...")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Printf("Error executing request: %v", err)
		return nil, nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	c.logger.Printf("Response status: %d %s", resp.StatusCode, resp.Status)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Printf("Error reading response body: %v", err)
		return nil, resp, fmt.Errorf("error reading response body: %w", err)
	}

	c.logger.Printf("Response body length: %d bytes", len(respBody))
	if len(respBody) > 0 {
		c.logger.Printf("Response body: %s", string(respBody))
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		c.logger.Printf("API error: %s (status %d)", string(respBody), resp.StatusCode)
		return respBody, resp, fmt.Errorf("API error: %s (status %d)", string(respBody), resp.StatusCode)
	}

	c.logger.Printf("File auth request successful")
	return respBody, resp, nil
}

// doWriteRequest performs an HTTP write request (POST, PATCH, DELETE) with rate limiting
func (c *Client) doWriteRequest(ctx context.Context, method, path string, body []byte, version int) ([]byte, *http.Response, error) {
	// Apply rate limiting
	if c.rateLimiter != nil {
		c.logger.Printf("Waiting for rate limiter...")
		if err := c.rateLimiter.Wait(ctx); err != nil {
			c.logger.Printf("Rate limiter error: %v", err)
			return nil, nil, fmt.Errorf("rate limiter error: %w", err)
		}
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/%s/%s%s",
		c.BaseURL,
		c.LibraryType,
		c.LibraryID,
		path,
	)

	c.logger.Printf("Making write request: %s %s", method, urlStr)

	// Create request
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
		c.logger.Printf("Request body: %s", string(body))
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, reqBody)
	if err != nil {
		c.logger.Printf("Error creating request: %v", err)
		return nil, nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	if c.APIKey != "" {
		req.Header.Set("Zotero-API-Key", c.APIKey)
		c.logger.Printf("API Key set")
	} else {
		c.logger.Printf("No API Key set")
	}
	req.Header.Set("Zotero-API-Version", "3")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set version header for concurrency control
	if version > 0 {
		req.Header.Set("If-Unmodified-Since-Version", strconv.Itoa(version))
		c.logger.Printf("If-Unmodified-Since-Version: %d", version)
	}

	// Execute request
	c.logger.Printf("Executing write request...")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Printf("Error executing request: %v", err)
		return nil, nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	c.logger.Printf("Response status: %d %s", resp.StatusCode, resp.Status)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Printf("Error reading response body: %v", err)
		return nil, resp, fmt.Errorf("error reading response body: %w", err)
	}

	c.logger.Printf("Response body length: %d bytes", len(respBody))
	if len(respBody) > 0 {
		c.logger.Printf("Response body: %s", string(respBody))
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		c.logger.Printf("API error: %s (status %d)", string(respBody), resp.StatusCode)
		return respBody, resp, fmt.Errorf("API error: %s (status %d)", string(respBody), resp.StatusCode)
	}

	c.logger.Printf("Write request successful")
	return respBody, resp, nil
}
