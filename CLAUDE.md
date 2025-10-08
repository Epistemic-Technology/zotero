# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go client library for the Zotero API that enables interaction with the Zotero Web API v3. The library is modeled after the Python pyzotero implementation and provides comprehensive read and write access to Zotero libraries, collections, items, searches, and tags.

## Commands

### Using Make (Recommended)
```bash
make help              # Show all available make targets
make build             # Build all binaries
make zotero-cli        # Build only the CLI tool
make test              # Run unit tests (default, fast)
make test-unit         # Run unit tests only (mock tests)
make test-integration  # Run integration tests (requires .env)
make test-all          # Run both unit and integration tests
make clean             # Remove build artifacts
```

### Testing

#### Unit Tests (Mock Tests)
Fast tests using mock HTTP servers and fixture data. No credentials required.

```bash
make test-unit                   # Run unit tests with make
go test ./zotero -v              # Run unit tests directly
go test ./zotero -run TestName   # Run specific unit test
go test ./zotero -race           # Run with race detector
go test ./zotero -cover          # Run with coverage
```

#### Integration Tests (Live/Local API)
Tests against real Zotero APIs. Requires credentials.

**Setup Option 1 - Using `.env` file (local development):**
1. Copy `.env.example` to `.env`
2. Add your credentials to `.env`:
   ```bash
   ZOTERO_API_KEY=your_api_key_here
   ZOTERO_LIBRARY_ID=your_library_id_here
   ZOTERO_LIBRARY_TYPE=user
   TEST_API_URL=https://api.zotero.org  # or http://localhost:23119 for local
   ```

**Setup Option 2 - Using shell environment (permanent setup):**
Add to `~/.zshrc` or `~/.bashrc`:
```bash
export ZOTERO_API_KEY=your_api_key_here
export ZOTERO_LIBRARY_ID=your_library_id_here
export ZOTERO_LIBRARY_TYPE=user
export TEST_API_URL=https://api.zotero.org
```

Get API key from https://www.zotero.org/settings/keys

**Run integration tests:**
```bash
make test-integration            # Run integration tests with make
go test ./tests -v               # Run integration tests directly
go test ./tests -run TestItems   # Run specific integration test
```

**Testing against local Zotero REST API:**
Set `TEST_API_URL=http://localhost:23119` in `.env` to test against Zotero desktop's local API instead of the live web API.

See `tests/README.md` for detailed integration testing documentation.

### Building
```bash
go build ./...                          # Build all packages
go build -o bin/zotero-cli ./cmd/zotero-cli  # Build CLI tool
```

### Development
```bash
go mod tidy                      # Clean up dependencies
go fmt ./...                     # Format code
go vet ./...                     # Run Go vet
```

### CLI Tool

The CLI tool uses the same environment variables as the integration tests. Set them once and use for both!

```bash
# Option 1: Set environment variables (recommended - works for CLI and tests)
export ZOTERO_API_KEY=your_key
export ZOTERO_LIBRARY_ID=your_library_id
export ZOTERO_LIBRARY_TYPE=user  # or group

# Then you can omit flags:
bin/zotero-cli items -limit 10
bin/zotero-cli items -itemtype journalArticle -limit 10
bin/zotero-cli items -itemtype "-annotation" -limit 20
bin/zotero-cli collections
bin/zotero-cli download -item ABC123 -path ./downloads

# Option 2: Use command-line flags
bin/zotero-cli items -library 12345 -key your_key -limit 10
bin/zotero-cli items -library 12345 -itemtype "book,journalArticle" -limit 20
bin/zotero-cli item -library 12345 -item ABC123
bin/zotero-cli collections -library 12345
bin/zotero-cli groups -user 12345
bin/zotero-cli download -library 12345 -item ABC123 -filename article.pdf -path ./downloads
```

## Architecture

### Core Structure

The library consists of the following main files in the `zotero/` package:

1. **zotero.go** - Client configuration and initialization
   - `Client` struct manages API connections with library ID, type (user/group), and API key
   - Functional options pattern via `ClientOption` for flexible configuration
   - Built-in rate limiting using `golang.org/x/time/rate` (default 1 request/second)
   - HTTP request handling with `doRequest()` method supporting context, rate limiting, and error handling
   - Configurable retry logic via `RetryConfig` (currently defined but not implemented)
   - Logger support for debugging API requests

2. **models.go** - Zotero API data models
   - `Item` and `ItemData` - Library items (books, articles, notes, attachments, etc.)
   - `Collection` and `CollectionData` - Item collections with hierarchical support
   - `Search` and `SearchData` - Saved searches with conditions
   - `Group` and `GroupMeta` - Group library information and metadata
   - `WriteResponse` and `FailedWrite` - Write operation responses with success/unchanged/failed status
   - `DeletedContent` - Tracking deleted resources (items, collections, searches, tags)
   - `TagsResponse` - Tag information with usage counts
   - `Creator`, `Tag`, `Relations` - Supporting structures for items
   - `Library`, `Links`, `Link`, `Meta` - Metadata structures used across resources

3. **read.go** - Read operations for the Zotero API
   - Query parameter construction via `QueryParams` struct
   - HTTP request handling with rate limiting and context support
   - Item operations: `Items()`, `Top()`, `Item()`, `Children()`, `Trash()`
   - Collection operations: `Collections()`, `CollectionsTop()`, `Collection()`, `CollectionsSub()`, `CollectionItems()`, `CollectionItemsTop()`
   - Search operations: `Searches()`, `Search()`
   - Tag operations: `Tags()`, `ItemTags()`, `CollectionTags()`
   - Group operations: `Groups()` (requires user library type)
   - File operations: `File()` - Download raw attachment content, `Dump()` - Download and save attachment to disk with auto-filename detection
   - Utility methods: `NumItems()`, `LastModifiedVersion()`, `Deleted()`

4. **write.go** - Write operations for the Zotero API
   - Item operations: `CreateItems()`, `UpdateItem()`, `UpdateItems()`, `DeleteItem()`, `DeleteItems()`
   - Collection operations: `CreateCollections()`, `UpdateCollection()`, `UpdateCollections()`, `DeleteCollection()`, `DeleteCollections()`
   - Search operations: `CreateSearches()`, `UpdateSearch()`, `DeleteSearch()`, `DeleteSearches()`
   - Tag operations: `AddTags()`, `DeleteTags()`
   - Attachment operations: `UploadAttachment()` - Multi-step file upload with authorization, upload to storage, and registration
   - All write operations support batch processing (up to 50 items per request)
   - Version-based concurrency control via `If-Unmodified-Since-Version` header
   - Returns `WriteResponse` for batch operations showing success/unchanged/failed items
   - Helper methods: `doWriteRequest()`, `doFileAuthRequest()` for handling write and file upload requests

5. **itemtypes.go** - Item type and creator type constants
   - String constants for common item types (book, journalArticle, webpage, etc.)
   - String constants for common creator types (author, editor, contributor, etc.)
   - Provides IDE autocomplete and type safety for most common use cases
   - Helper functions: `IsExcludeFilter()`, `WithoutExcludePrefix()`
   - Users can still use raw strings for any item type not listed as a constant

6. **schema.go** - Dynamic schema fetching from Zotero API
   - `ItemTypes()` - Fetch all available item types with localization
   - `ItemFields()` - Fetch all available fields
   - `ItemTypeFields()` - Fetch valid fields for a specific item type
   - `ItemTypeCreatorTypes()` - Fetch valid creator types for a specific item type
   - `CreatorFields()` - Fetch localized creator field names
   - `NewItemTemplate()` - Get a template for creating new items (recommended before creating items)
   - All methods support optional locale parameter for internationalization

### Test Files

**Unit Tests (`zotero/` package):**
- **read_test.go** - Unit tests for read operations using mock HTTP servers
- **write_test.go** - Unit tests for write operations using mock HTTP servers
- **models_test.go** - Tests for data model serialization/deserialization
- **itemtypes_test.go** - Tests for item type constants and helper functions
- **schema_test.go** - Tests for schema fetching operations
- **zotero_test.go** - Tests for client configuration and initialization

**Integration Tests (`tests/` package):**
- **integration_test.go** - Integration tests for read operations against live/local API
- **write_integration_test.go** - Integration tests for write operations against live/local API

### Client Configuration

The `NewClient()` function accepts a library ID and type, plus optional configuration via:
- `WithAPIKey()` - Authentication (required for private libraries)
- `WithBaseURL()` - Custom API endpoint (default: https://api.zotero.org)
- `WithLocale()` - Localization settings (default: en-US)
- `WithTimeout()` - HTTP request timeout (default: 30 seconds)
- `WithRateLimit()` - API rate limiting (default: 1 second between requests)
- `WithRetry()` - Retry configuration for failed requests (not yet implemented)
- `WithHTTPClient()` - Custom HTTP client
- `WithPreserveJSON()` - JSON field ordering (not yet implemented)
- `WithLogger()` - Custom logger for debugging API requests

### Library Types

Two library types are supported via the `LibraryType` enum:
- `LibraryTypeUser` - Personal user libraries
- `LibraryTypeGroup` - Shared group libraries

### Query Parameters

API requests can be customized with `QueryParams`:
- `Limit` - Maximum number of results (default 100)
- `Start` - Starting index for pagination
- `Sort` - Sort field (dateAdded, dateModified, title, creator, itemType, etc.)
- `Format` - Response format (atom, bib, json, keys, versions, etc.)
- `Include` - Additional data (data, bib, citation, etc.)
- `Style` - Citation style for bibliographic formats
- `Q` - Quick search query
- `QMode` - Quick search mode (titleCreatorYear, everything)
- `Tag` - Filter by tag(s)
- `ItemKey` - Filter by item key(s)
- `ItemType` - Filter by item type(s); prefix with "-" to exclude (e.g., []string{"journalArticle"} or []string{"-annotation"})
- `Since` - Return only objects modified since version
- `Extra` - Additional query parameters

#### Item Type Filtering Examples

The library provides constants for common item types with IDE autocomplete support:

```go
import "github.com/Epistemic-Technology/zotero/zotero"

// Using constants (recommended - provides IDE autocomplete)
items, err := client.Items(ctx, &zotero.QueryParams{
    ItemType: []string{zotero.ItemTypeJournalArticle},
    Limit:    25,
})

// Exclude annotations using constant
items, err := client.Items(ctx, &zotero.QueryParams{
    ItemType: []string{"-" + zotero.ItemTypeAnnotation},
    Limit:    50,
})

// Multiple filters: books and journal articles, excluding annotations
items, err := client.Items(ctx, &zotero.QueryParams{
    ItemType: []string{
        zotero.ItemTypeBook,
        zotero.ItemTypeJournalArticle,
        "-" + zotero.ItemTypeAnnotation,
    },
})

// You can still use raw strings for item types not available as constants
items, err := client.Items(ctx, &zotero.QueryParams{
    ItemType: []string{"customItemType"},
})
```

Available item type constants include: `ItemTypeBook`, `ItemTypeJournalArticle`, `ItemTypeWebpage`, `ItemTypeAttachment`, `ItemTypeNote`, `ItemTypeAnnotation`, `ItemTypeConferencePaper`, `ItemTypeThesis`, `ItemTypeReport`, `ItemTypeBlogPost`, `ItemTypePodcast`, `ItemTypeVideoRecording`, and many more (see `zotero/itemtypes.go` for the complete list).

### Data Model Design

Items use a flexible structure where common fields are explicitly defined in `ItemData`, while item-type-specific fields can be stored in the `Extra map[string]any` field to accommodate Zotero's varied item types (books, articles, web pages, etc.).

Relations between items use the `Relations` struct with Dublin Core and OWL predicates for semantic relationships.

### Write Operations

All write operations require an API key with write permissions. The library provides comprehensive write support modeled after pyzotero:

#### Creating Items

```go
// Get a template for the item type (recommended)
template, err := client.NewItemTemplate(ctx, zotero.ItemTypeBook)
if err != nil {
    log.Fatal(err)
}

// Populate the template
item := zotero.Item{
    Data: zotero.ItemData{
        ItemType: zotero.ItemTypeBook,
        Title:    "The Go Programming Language",
        Creators: []zotero.Creator{
            {CreatorType: zotero.CreatorTypeAuthor, FirstName: "Alan", LastName: "Donovan"},
            {CreatorType: zotero.CreatorTypeAuthor, FirstName: "Brian", LastName: "Kernighan"},
        },
    },
}

// Create the item
resp, err := client.CreateItems(ctx, []zotero.Item{item})
if err != nil {
    log.Fatal(err)
}

// Check response for success/failures
for idx, key := range resp.Success {
    fmt.Printf("Created item %s at index %s\n", key, idx)
}
for idx, failure := range resp.Failed {
    fmt.Printf("Failed to create item at index %s: %s\n", idx, failure.Message)
}
```

#### Updating Items

```go
// Fetch the current item (to get version number)
item, err := client.Item(ctx, "ABCD1234", nil)
if err != nil {
    log.Fatal(err)
}

// Modify the item
item.Data.Title = "Updated Title"

// Update single item
err = client.UpdateItem(ctx, item)
if err != nil {
    log.Fatal(err)
}

// Or update multiple items at once (up to 50)
items := []zotero.Item{item1, item2, item3}
resp, err := client.UpdateItems(ctx, items)
```

#### Deleting Items

```go
// Delete a single item (requires version for concurrency control)
err := client.DeleteItem(ctx, "ABCD1234", version)
if err != nil {
    log.Fatal(err)
}

// Delete multiple items at once (up to 50)
itemKeys := []string{"ABCD1234", "EFGH5678"}
err = client.DeleteItems(ctx, itemKeys, version)
```

#### Collections

```go
// Create a collection
collection := zotero.Collection{
    Data: zotero.CollectionData{
        Name: "My Research Papers",
        ParentCollection: "", // empty for top-level
    },
}
resp, err := client.CreateCollections(ctx, []zotero.Collection{collection})

// Update a collection
collection.Data.Name = "Updated Name"
err = client.UpdateCollection(ctx, &collection)

// Delete a collection
err = client.DeleteCollection(ctx, "COLL1234", version)
```

#### Searches

```go
// Create a saved search
search := zotero.Search{
    Data: zotero.SearchData{
        Name: "Recent Articles",
        Conditions: []zotero.SearchCondition{
            {Condition: "itemType", Operator: "is", Value: "journalArticle"},
            {Condition: "date", Operator: "isInTheLast", Value: "30 days"},
        },
    },
}
resp, err := client.CreateSearches(ctx, []zotero.Search{search})

// Update a search
err = client.UpdateSearch(ctx, &search)

// Delete a search
err = client.DeleteSearch(ctx, "SRCH1234", version)
```

#### Tags

```go
// Add tags to an item (convenience method that fetches, modifies, and updates)
err := client.AddTags(ctx, "ABCD1234", "important", "to-read", "golang")

// Delete tags from the library (removes from all items)
err = client.DeleteTags(ctx, version, "obsolete", "old-tag")
```

#### Attachments

##### Uploading Files

```go
// Upload a file as an attachment to a parent item
attachment, err := client.UploadAttachment(ctx, parentItemKey, "/path/to/file.pdf", "file.pdf", "application/pdf")
if err != nil {
    log.Fatal(err)
}

// Create a standalone attachment (no parent)
attachment, err := client.UploadAttachment(ctx, "", "/path/to/document.pdf", "document.pdf", "application/pdf")

// The method handles the complete multi-step upload process:
// 1. Creates an attachment item with linkMode "imported_file"
// 2. Requests upload authorization from Zotero
// 3. Uploads the file to cloud storage (S3)
// 4. Registers the upload with Zotero
// 5. Returns the completed attachment item
```

##### Downloading Files

```go
// Download raw file content
fileContent, err := client.File(ctx, "ABCD1234")
if err != nil {
    log.Fatal(err)
}
// fileContent is a []byte containing the file data

// Save file to disk with automatic filename detection
fullPath, err := client.Dump(ctx, "ABCD1234", "", "")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("File saved to: %s\n", fullPath)

// Save file with custom filename and path
fullPath, err := client.Dump(ctx, "ABCD1234", "my-article.pdf", "/path/to/downloads")
if err != nil {
    log.Fatal(err)
}
// File saved to: /path/to/downloads/my-article.pdf

// Dump() automatically:
// 1. Fetches item metadata to determine filename (if not provided)
// 2. Downloads the file content
// 3. Writes to disk
// 4. Returns the full path to the saved file
```

#### Batch Operations

All write operations support batch processing with up to 50 items per request:

```go
// Create up to 50 items at once
items := make([]zotero.Item, 50)
for i := range items {
    items[i] = zotero.Item{
        Data: zotero.ItemData{
            ItemType: zotero.ItemTypeBook,
            Title:    fmt.Sprintf("Book %d", i),
        },
    }
}
resp, err := client.CreateItems(ctx, items)

// Response includes success, unchanged, and failed items
fmt.Printf("Success: %d, Unchanged: %d, Failed: %d\n", 
    len(resp.Success), len(resp.Unchanged), len(resp.Failed))
```

#### Concurrency Control

All update and delete operations require version information to prevent conflicts:

```go
// Always fetch the current version before updating
item, err := client.Item(ctx, "ABCD1234", nil)
if err != nil {
    log.Fatal(err)
}

// Version is automatically included from the item
err = client.UpdateItem(ctx, item)

// For deletes, you must provide the version explicitly
err = client.DeleteItem(ctx, "ABCD1234", item.Version)
```

The API returns a 412 Precondition Failed error if the version doesn't match, indicating the item was modified by another client.

### Schema Fetching

For advanced use cases (e.g., validation, UI generation, supporting new item types), you can fetch the current Zotero schema dynamically:

```go
// Fetch all available item types
itemTypes, err := client.ItemTypes(ctx, "en-US")
for _, it := range itemTypes {
    fmt.Printf("%s: %s\n", it.ItemType, it.Localized)
}

// Fetch valid fields for a specific item type
fields, err := client.ItemTypeFields(ctx, zotero.ItemTypeBook, "")
for _, field := range fields {
    fmt.Printf("%s: %s\n", field.Field, field.Localized)
}

// Fetch valid creator types for a specific item type
creatorTypes, err := client.ItemTypeCreatorTypes(ctx, zotero.ItemTypeJournalArticle, "")
for _, ct := range creatorTypes {
    fmt.Printf("%s: %s\n", ct.CreatorType, ct.Localized)
}

// Get a template for creating new items (useful for future write operations)
template, err := client.NewItemTemplate(ctx, zotero.ItemTypeBook)
// template is a map[string]any with all fields for the item type
```

Schema methods support optional locale parameters (e.g., "en-US", "de-DE", "fr-FR") for internationalization. The Zotero API recommends caching schema data for about an hour.

### CLI Tool

The `cmd/zotero-cli` package provides a command-line interface for interacting with the Zotero API:

**Commands**:
- `items` - List items in a library with pagination support (limit, start, itemtype filtering)
- `item` - Get a specific item by key
- `collections` - List all collections in a library
- `groups` - List groups for a user
- `create` - Create a new item with optional file attachment
- `upload` - Upload a file attachment to an item
- `download` - Download a file attachment from an item

**Environment variable support**:
- `ZOTERO_API_KEY` - API key for authentication
- `ZOTERO_LIBRARY_ID` - Library ID (default for commands)
- `ZOTERO_LIBRARY_TYPE` - Library type: user or group (default: user)

**Features**:
- Tabular output for lists (items, collections, groups) with formatted columns
- Detailed output for single item view
- Item type filtering with `-itemtype` flag (supports comma-separated list and exclusion with "-" prefix)
- Verbose logging flag (`-v`) for debugging API requests
- Command-line flags override environment variables
- Built on top of the core `zotero` package

## Testing Strategy

The project uses a comprehensive two-tier testing approach:

### Unit Tests (`zotero/` package)
- **Location**: `zotero/*_test.go` (6 test files)
- **Type**: White-box tests using mock HTTP servers
- **Fixtures**: JSON test data in `zotero/testdata/`
- **Speed**: Fast, no external dependencies
- **Run**: `make test-unit` or `go test ./zotero -v`
- **Coverage**:
  - `read_test.go` - All read operations (Items, Collections, Searches, Tags, Groups, File downloads)
  - `write_test.go` - All write operations (Create, Update, Delete for items, collections, searches, tags, file uploads)
  - `models_test.go` - Data model serialization/deserialization
  - `itemtypes_test.go` - Item type constants and helper functions
  - `schema_test.go` - Schema fetching operations
  - `zotero_test.go` - Client configuration and initialization
- **Purpose**: Test internal logic, data models, query building, error handling, context cancellation

### Integration Tests (`tests/` package)
- **Location**: `tests/integration_test.go`, `tests/write_integration_test.go`
- **Type**: Black-box tests against real Zotero APIs
- **Requirements**: API credentials in `.env` file or environment variables (write operations require API key with write permissions)
- **APIs Supported**: Live Zotero Web API (`https://api.zotero.org`) or local Zotero desktop REST API (`http://localhost:23119`)
- **Run**: `make test-integration` or `go test ./tests -v`
- **Coverage**:
  - `integration_test.go` - Read operations against live API (pagination, sorting, filtering, item type filtering)
  - `write_integration_test.go` - Write operations with automatic cleanup:
    - Single and batch item create/update/delete
    - Collection operations including nested collections
    - Saved search operations
    - Tag operations
    - File upload and download
    - Version-based concurrency control validation
- **Features**:
  - Auto-skip if credentials not available
  - Switch between live and local API via `TEST_API_URL` environment variable
  - Automatic cleanup of test resources using deferred deletion
  - Comprehensive error handling and validation
  - See `tests/README.md` for detailed documentation

## Current Status

### Implemented Features
- ✅ Complete read API support for items, collections, searches, tags, and groups
- ✅ Complete write API support:
  - ✅ Create, update, delete items (single and batch operations)
  - ✅ Create, update, delete collections (single and batch operations)
  - ✅ Create, update, delete saved searches (single and batch operations)
  - ✅ Add and delete tags
  - ✅ Upload attachments (multi-step file upload process)
  - ✅ Download attachments (File() and Dump() methods)
  - ✅ Version-based concurrency control (412 Precondition Failed on version mismatch)
  - ✅ Batch operations (up to 50 items per request)
  - ✅ WriteResponse with success/unchanged/failed tracking
- ✅ Rate limiting and timeout configuration
- ✅ Context support for all API calls
- ✅ Query parameter support (pagination, sorting, filtering, formats)
- ✅ CLI tool with environment variable support
- ✅ Logger integration for debugging
- ✅ Unit tests with mock HTTP servers and fixtures (read and write operations)
- ✅ Integration tests for live/local API testing (read and write operations)
- ✅ Schema fetching with localization support
- ✅ Item type and creator type constants for IDE autocomplete

### Not Yet Implemented
- ❌ Retry logic for failed requests (RetryConfig defined but not used)
- ❌ JSON field ordering preservation (preserveJSON flag defined but not used)
- ❌ Full-text search
- ❌ Adding items to collections via write API
- ❌ Removing items from collections via write API

## External References

- [Zotero Web API v3 Documentation](https://www.zotero.org/support/dev/web_api/v3/start)
- [Pyzotero documentation](https://pyzotero.readthedocs.io/en/latest/) - Reference implementation
