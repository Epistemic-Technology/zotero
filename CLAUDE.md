# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go client library for the Zotero API that enables interaction with the Zotero Web API v3. The library is modeled after the Python pyzotero implementation and provides comprehensive read-only access to Zotero libraries, collections, items, searches, and tags.

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

# Option 2: Use command-line flags
bin/zotero-cli items -library 12345 -key your_key -limit 10
bin/zotero-cli items -library 12345 -itemtype "book,journalArticle" -limit 20
bin/zotero-cli item -library 12345 -item ABC123
bin/zotero-cli collections -library 12345
bin/zotero-cli groups -user 12345
```

## Architecture

### Core Structure

The library consists of three main files in the `zotero/` package:

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
   - `WriteResponse` and `FailedWrite` - Write operation responses (models only, write operations not yet implemented)
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
   - Utility methods: `NumItems()`, `LastModifiedVersion()`, `Deleted()`

4. **itemtypes.go** - Item type and creator type constants
   - String constants for common item types (book, journalArticle, webpage, etc.)
   - String constants for common creator types (author, editor, contributor, etc.)
   - Provides IDE autocomplete and type safety for most common use cases
   - Helper functions: `IsExcludeFilter()`, `WithoutExcludePrefix()`
   - Users can still use raw strings for any item type not listed as a constant

5. **schema.go** - Dynamic schema fetching from Zotero API
   - `ItemTypes()` - Fetch all available item types with localization
   - `ItemFields()` - Fetch all available fields
   - `ItemTypeFields()` - Fetch valid fields for a specific item type
   - `ItemTypeCreatorTypes()` - Fetch valid creator types for a specific item type
   - `CreatorFields()` - Fetch localized creator field names
   - `NewItemTemplate()` - Get a template for creating new items (useful for future write operations)
   - All methods support optional locale parameter for internationalization

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
- **Commands**:
  - `items` - List items in a library with pagination support (limit, start, itemtype filtering)
  - `item` - Get a specific item by key
  - `collections` - List all collections in a library
  - `groups` - List groups for a user
- **Environment variable support**:
  - `ZOTERO_API_KEY` - API key for authentication
  - `ZOTERO_LIBRARY_ID` - Library ID (default for commands)
  - `ZOTERO_LIBRARY_TYPE` - Library type: user or group (default: user)
- **Features**:
  - JSON output formatting with indentation
  - Item type filtering with `-itemtype` flag (supports comma-separated list and exclusion with "-" prefix)
  - Verbose logging flag (`-v`) for debugging
  - Command-line flags override environment variables
- Built on top of the core `zotero` package

## Testing Strategy

The project has two types of tests:

### Unit Tests (`zotero/` package)
- **Location**: `zotero/*_test.go`
- **Type**: White-box tests using mock HTTP servers
- **Fixtures**: JSON test data in `zotero/testdata/`
- **Speed**: Fast, no external dependencies
- **Run**: `make test-unit` or `go test ./zotero`
- **Purpose**: Test internal logic, data models, query building, error handling

### Integration Tests (`tests/` package)
- **Location**: `tests/integration_test.go`
- **Type**: Black-box tests against real Zotero APIs
- **Requirements**: API credentials in `.env` file
- **APIs**: Live Zotero Web API or local Zotero desktop REST API
- **Run**: `make test-integration` or `go test ./tests`
- **Purpose**: Verify end-to-end functionality with real API
- **Features**:
  - Auto-skip if credentials not available
  - Switch between live and local API via `TEST_API_URL`
  - Test all read operations, pagination, sorting, filtering
  - See `tests/README.md` for detailed documentation

## Current Status

### Implemented Features
- ✅ Complete read-only API support for items, collections, searches, tags, and groups
- ✅ Rate limiting and timeout configuration
- ✅ Context support for all API calls
- ✅ Query parameter support (pagination, sorting, filtering, formats)
- ✅ CLI tool with environment variable support
- ✅ Logger integration for debugging
- ✅ Unit tests with mock HTTP servers and fixtures
- ✅ Integration tests for live/local API testing

### Not Yet Implemented
- ❌ Write operations (create, update, delete items/collections)
- ❌ Retry logic for failed requests (RetryConfig defined but not used)
- ❌ JSON field ordering preservation (preserveJSON flag defined but not used)
- ❌ Attachment upload/download
- ❌ Full-text search

## External References

- [Zotero Web API v3 Documentation](https://www.zotero.org/support/dev/web_api/v3/start)
- [Pyzotero documentation](https://pyzotero.readthedocs.io/en/latest/) - Reference implementation
