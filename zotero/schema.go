package zotero

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SchemaItemType represents an item type from the Zotero schema
type SchemaItemType struct {
	ItemType  string `json:"itemType"`
	Localized string `json:"localized,omitempty"`
}

// SchemaField represents a field from the Zotero schema
type SchemaField struct {
	Field     string `json:"field"`
	Localized string `json:"localized,omitempty"`
}

// SchemaCreatorType represents a creator type from the Zotero schema
type SchemaCreatorType struct {
	CreatorType string `json:"creatorType"`
	Localized   string `json:"localized,omitempty"`
}

// SchemaItemTypeFields represents fields for a specific item type
type SchemaItemTypeFields struct {
	ItemType string        `json:"itemType"`
	Fields   []SchemaField `json:"fields"`
}

// SchemaItemTypeCreatorTypes represents creator types for a specific item type
type SchemaItemTypeCreatorTypes struct {
	ItemType     string              `json:"itemType"`
	CreatorTypes []SchemaCreatorType `json:"creatorTypes"`
}

// NewItemTemplate represents a template for creating a new item
type NewItemTemplate map[string]any

// ItemTypes retrieves all available item types from the Zotero schema.
// The locale parameter is optional (e.g., "en-US", "de-DE"). If empty, defaults to client's locale.
func (c *Client) ItemTypes(ctx context.Context, locale string) ([]SchemaItemType, error) {
	params := &QueryParams{}
	if locale != "" {
		params.Extra = map[string]string{"locale": locale}
	}

	body, _, err := c.doRequest(ctx, http.MethodGet, "/itemTypes", params)
	if err != nil {
		return nil, err
	}

	var itemTypes []SchemaItemType
	if err := json.Unmarshal(body, &itemTypes); err != nil {
		return nil, fmt.Errorf("error unmarshaling item types: %w", err)
	}

	return itemTypes, nil
}

// ItemFields retrieves all available item fields from the Zotero schema.
// The locale parameter is optional (e.g., "en-US", "de-DE"). If empty, defaults to client's locale.
func (c *Client) ItemFields(ctx context.Context, locale string) ([]SchemaField, error) {
	params := &QueryParams{}
	if locale != "" {
		params.Extra = map[string]string{"locale": locale}
	}

	body, _, err := c.doRequest(ctx, http.MethodGet, "/itemFields", params)
	if err != nil {
		return nil, err
	}

	var fields []SchemaField
	if err := json.Unmarshal(body, &fields); err != nil {
		return nil, fmt.Errorf("error unmarshaling item fields: %w", err)
	}

	return fields, nil
}

// ItemTypeFields retrieves valid fields for a specific item type.
// The locale parameter is optional (e.g., "en-US", "de-DE"). If empty, defaults to client's locale.
func (c *Client) ItemTypeFields(ctx context.Context, itemType string, locale string) ([]SchemaField, error) {
	path := fmt.Sprintf("/itemTypeFields?itemType=%s", itemType)
	params := &QueryParams{}
	if locale != "" {
		params.Extra = map[string]string{"locale": locale}
	}

	body, _, err := c.doRequest(ctx, http.MethodGet, path, params)
	if err != nil {
		return nil, err
	}

	var fields []SchemaField
	if err := json.Unmarshal(body, &fields); err != nil {
		return nil, fmt.Errorf("error unmarshaling item type fields: %w", err)
	}

	return fields, nil
}

// ItemTypeCreatorTypes retrieves valid creator types for a specific item type.
// The locale parameter is optional (e.g., "en-US", "de-DE"). If empty, defaults to client's locale.
func (c *Client) ItemTypeCreatorTypes(ctx context.Context, itemType string, locale string) ([]SchemaCreatorType, error) {
	path := fmt.Sprintf("/itemTypeCreatorTypes?itemType=%s", itemType)
	params := &QueryParams{}
	if locale != "" {
		params.Extra = map[string]string{"locale": locale}
	}

	body, _, err := c.doRequest(ctx, http.MethodGet, path, params)
	if err != nil {
		return nil, err
	}

	var creatorTypes []SchemaCreatorType
	if err := json.Unmarshal(body, &creatorTypes); err != nil {
		return nil, fmt.Errorf("error unmarshaling creator types: %w", err)
	}

	return creatorTypes, nil
}

// CreatorFields retrieves localized creator field names (firstName, lastName, name).
// The locale parameter is optional (e.g., "en-US", "de-DE"). If empty, defaults to client's locale.
func (c *Client) CreatorFields(ctx context.Context, locale string) ([]SchemaField, error) {
	params := &QueryParams{}
	if locale != "" {
		params.Extra = map[string]string{"locale": locale}
	}

	body, _, err := c.doRequest(ctx, http.MethodGet, "/creatorFields", params)
	if err != nil {
		return nil, err
	}

	var fields []SchemaField
	if err := json.Unmarshal(body, &fields); err != nil {
		return nil, fmt.Errorf("error unmarshaling creator fields: %w", err)
	}

	return fields, nil
}

// NewItemTemplate retrieves a template for creating a new item of the specified type.
// The template includes all valid fields for the item type with empty/default values.
// This is useful when implementing write operations to ensure all required fields are present.
func (c *Client) NewItemTemplate(ctx context.Context, itemType string) (NewItemTemplate, error) {
	path := fmt.Sprintf("/items/new?itemType=%s", itemType)

	body, _, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var template NewItemTemplate
	if err := json.Unmarshal(body, &template); err != nil {
		return nil, fmt.Errorf("error unmarshaling item template: %w", err)
	}

	return template, nil
}
