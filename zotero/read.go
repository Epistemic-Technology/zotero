package zotero

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// QueryParams represents optional parameters for API requests
type QueryParams struct {
	Limit    int               // Maximum number of results (default 100)
	Start    int               // Starting index for results
	Sort     string            // Field to sort by (dateAdded, dateModified, title, creator, itemType, etc.)
	Format   string            // Response format (atom, bib, json, keys, versions, etc.)
	Include  string            // Additional data to include (data, bib, citation, etc.)
	Style    string            // Citation style for bib/citation formats
	Q        string            // Quick search query
	QMode    string            // Quick search mode (titleCreatorYear, everything)
	Tag      []string          // Filter by tag(s)
	ItemKey  []string          // Filter by item key(s)
	ItemType []string          // Filter by item type(s); prefix with "-" to exclude (e.g., "-annotation")
	Since    int               // Return only objects modified since version
	Extra    map[string]string // Additional query parameters
}

// Items retrieves all library items
func (c *Client) Items(ctx context.Context, params *QueryParams) ([]Item, error) {
	body, _, err := c.doRequest(ctx, http.MethodGet, "/items", params)
	if err != nil {
		return nil, err
	}

	var items []Item
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("error unmarshaling items: %w", err)
	}

	return items, nil
}

// Top retrieves top-level library items (no parent items)
func (c *Client) Top(ctx context.Context, params *QueryParams) ([]Item, error) {
	body, _, err := c.doRequest(ctx, http.MethodGet, "/items/top", params)
	if err != nil {
		return nil, err
	}

	var items []Item
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("error unmarshaling items: %w", err)
	}

	return items, nil
}

// Item retrieves a specific item by key
func (c *Client) Item(ctx context.Context, itemKey string, params *QueryParams) (*Item, error) {
	path := fmt.Sprintf("/items/%s", itemKey)
	body, _, err := c.doRequest(ctx, http.MethodGet, path, params)
	if err != nil {
		return nil, err
	}

	var item Item
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("error unmarshaling item: %w", err)
	}

	return &item, nil
}

// Children retrieves child items of a specific item
func (c *Client) Children(ctx context.Context, itemKey string, params *QueryParams) ([]Item, error) {
	path := fmt.Sprintf("/items/%s/children", itemKey)
	body, _, err := c.doRequest(ctx, http.MethodGet, path, params)
	if err != nil {
		return nil, err
	}

	var items []Item
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("error unmarshaling items: %w", err)
	}

	return items, nil
}

// Trash retrieves items in the trash
func (c *Client) Trash(ctx context.Context, params *QueryParams) ([]Item, error) {
	body, _, err := c.doRequest(ctx, http.MethodGet, "/items/trash", params)
	if err != nil {
		return nil, err
	}

	var items []Item
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("error unmarshaling items: %w", err)
	}

	return items, nil
}

// Collections retrieves all library collections
func (c *Client) Collections(ctx context.Context, params *QueryParams) ([]Collection, error) {
	body, _, err := c.doRequest(ctx, http.MethodGet, "/collections", params)
	if err != nil {
		return nil, err
	}

	var collections []Collection
	if err := json.Unmarshal(body, &collections); err != nil {
		return nil, fmt.Errorf("error unmarshaling collections: %w", err)
	}

	return collections, nil
}

// CollectionsTop retrieves top-level collections
func (c *Client) CollectionsTop(ctx context.Context, params *QueryParams) ([]Collection, error) {
	body, _, err := c.doRequest(ctx, http.MethodGet, "/collections/top", params)
	if err != nil {
		return nil, err
	}

	var collections []Collection
	if err := json.Unmarshal(body, &collections); err != nil {
		return nil, fmt.Errorf("error unmarshaling collections: %w", err)
	}

	return collections, nil
}

// Collection retrieves a specific collection by key
func (c *Client) Collection(ctx context.Context, collectionKey string, params *QueryParams) (*Collection, error) {
	path := fmt.Sprintf("/collections/%s", collectionKey)
	body, _, err := c.doRequest(ctx, http.MethodGet, path, params)
	if err != nil {
		return nil, err
	}

	var collection Collection
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, fmt.Errorf("error unmarshaling collection: %w", err)
	}

	return &collection, nil
}

// CollectionsSub retrieves subcollections of a specific collection
func (c *Client) CollectionsSub(ctx context.Context, collectionKey string, params *QueryParams) ([]Collection, error) {
	path := fmt.Sprintf("/collections/%s/collections", collectionKey)
	body, _, err := c.doRequest(ctx, http.MethodGet, path, params)
	if err != nil {
		return nil, err
	}

	var collections []Collection
	if err := json.Unmarshal(body, &collections); err != nil {
		return nil, fmt.Errorf("error unmarshaling collections: %w", err)
	}

	return collections, nil
}

// CollectionItems retrieves items from a specific collection
func (c *Client) CollectionItems(ctx context.Context, collectionKey string, params *QueryParams) ([]Item, error) {
	path := fmt.Sprintf("/collections/%s/items", collectionKey)
	body, _, err := c.doRequest(ctx, http.MethodGet, path, params)
	if err != nil {
		return nil, err
	}

	var items []Item
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("error unmarshaling items: %w", err)
	}

	return items, nil
}

// CollectionItemsTop retrieves top-level items from a specific collection
func (c *Client) CollectionItemsTop(ctx context.Context, collectionKey string, params *QueryParams) ([]Item, error) {
	path := fmt.Sprintf("/collections/%s/items/top", collectionKey)
	body, _, err := c.doRequest(ctx, http.MethodGet, path, params)
	if err != nil {
		return nil, err
	}

	var items []Item
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("error unmarshaling items: %w", err)
	}

	return items, nil
}

// Searches retrieves all saved searches
func (c *Client) Searches(ctx context.Context, params *QueryParams) ([]Search, error) {
	body, _, err := c.doRequest(ctx, http.MethodGet, "/searches", params)
	if err != nil {
		return nil, err
	}

	var searches []Search
	if err := json.Unmarshal(body, &searches); err != nil {
		return nil, fmt.Errorf("error unmarshaling searches: %w", err)
	}

	return searches, nil
}

// Search retrieves a specific saved search by key
func (c *Client) Search(ctx context.Context, searchKey string, params *QueryParams) (*Search, error) {
	path := fmt.Sprintf("/searches/%s", searchKey)
	body, _, err := c.doRequest(ctx, http.MethodGet, path, params)
	if err != nil {
		return nil, err
	}

	var search Search
	if err := json.Unmarshal(body, &search); err != nil {
		return nil, fmt.Errorf("error unmarshaling search: %w", err)
	}

	return &search, nil
}

// TagsResponse represents the response from the tags endpoint
type TagsResponse struct {
	Tag      string `json:"tag"`
	NumItems int    `json:"numItems,omitempty"`
	Type     int    `json:"type,omitempty"`
	Meta     Meta   `json:"meta,omitempty"`
	Links    Links  `json:"links,omitempty"`
}

// Tags retrieves all library tags
func (c *Client) Tags(ctx context.Context, params *QueryParams) ([]TagsResponse, error) {
	body, _, err := c.doRequest(ctx, http.MethodGet, "/tags", params)
	if err != nil {
		return nil, err
	}

	var tags []TagsResponse
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, fmt.Errorf("error unmarshaling tags: %w", err)
	}

	return tags, nil
}

// ItemTags retrieves tags for a specific item
func (c *Client) ItemTags(ctx context.Context, itemKey string, params *QueryParams) ([]Tag, error) {
	path := fmt.Sprintf("/items/%s/tags", itemKey)
	body, _, err := c.doRequest(ctx, http.MethodGet, path, params)
	if err != nil {
		return nil, err
	}

	var tags []Tag
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, fmt.Errorf("error unmarshaling tags: %w", err)
	}

	return tags, nil
}

// CollectionTags retrieves tags for items in a specific collection
func (c *Client) CollectionTags(ctx context.Context, collectionKey string, params *QueryParams) ([]TagsResponse, error) {
	path := fmt.Sprintf("/collections/%s/tags", collectionKey)
	body, _, err := c.doRequest(ctx, http.MethodGet, path, params)
	if err != nil {
		return nil, err
	}

	var tags []TagsResponse
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, fmt.Errorf("error unmarshaling tags: %w", err)
	}

	return tags, nil
}

// Groups retrieves groups the current user belongs to (requires user library type)
func (c *Client) Groups(ctx context.Context, params *QueryParams) ([]Group, error) {
	if c.LibraryType != LibraryTypeUser {
		return nil, fmt.Errorf("groups() requires user library type")
	}

	// Groups endpoint doesn't use library type/ID prefix
	urlStr := fmt.Sprintf("%s/users/%s/groups%s",
		c.BaseURL,
		c.LibraryID,
		c.buildQueryString(params),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if c.APIKey != "" {
		req.Header.Set("Zotero-API-Key", c.APIKey)
	}
	req.Header.Set("Zotero-API-Version", "3")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s (status %d)", string(body), resp.StatusCode)
	}

	var groups []Group
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, fmt.Errorf("error unmarshaling groups: %w", err)
	}

	return groups, nil
}

// NumItems returns the total count of library items
func (c *Client) NumItems(ctx context.Context) (int, error) {
	params := &QueryParams{
		Limit:  1,
		Format: "json",
	}

	_, resp, err := c.doRequest(ctx, http.MethodGet, "/items", params)
	if err != nil {
		return 0, err
	}

	totalResults := resp.Header.Get("Total-Results")
	if totalResults == "" {
		return 0, fmt.Errorf("Total-Results header not found")
	}

	count, err := strconv.Atoi(totalResults)
	if err != nil {
		return 0, fmt.Errorf("error parsing Total-Results: %w", err)
	}

	return count, nil
}

// LastModifiedVersion returns the library's last modified version
func (c *Client) LastModifiedVersion(ctx context.Context) (int, error) {
	_, resp, err := c.doRequest(ctx, http.MethodGet, "/items", &QueryParams{Limit: 1})
	if err != nil {
		return 0, err
	}

	version := resp.Header.Get("Last-Modified-Version")
	if version == "" {
		return 0, fmt.Errorf("Last-Modified-Version header not found")
	}

	v, err := strconv.Atoi(version)
	if err != nil {
		return 0, fmt.Errorf("error parsing Last-Modified-Version: %w", err)
	}

	return v, nil
}

// Deleted retrieves deleted content since a specific version
func (c *Client) Deleted(ctx context.Context, since int) (*DeletedContent, error) {
	params := &QueryParams{
		Since: since,
	}

	body, _, err := c.doRequest(ctx, http.MethodGet, "/deleted", params)
	if err != nil {
		return nil, err
	}

	var deleted DeletedContent
	if err := json.Unmarshal(body, &deleted); err != nil {
		return nil, fmt.Errorf("error unmarshaling deleted content: %w", err)
	}

	return &deleted, nil
}
