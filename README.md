# zotero

A Go client library for the Zotero API that enables comprehensive read and write access to Zotero libraries, collections, items, searches, and tags. This library allows for interaction with both the Zotero Web API v3 and the Zotero desktop application through the local REST API.

## Features

- ✅ **Complete Read API**: Items, collections, searches, tags, groups, and file downloads
- ✅ **Complete Write API**: Create, update, and delete operations with batch support (up to 50 items)
- ✅ **File Operations**: Upload and download attachments with multi-step upload support
- ✅ **Rate Limiting**: Built-in rate limiting and timeout configuration
- ✅ **Context Support**: Full context.Context support for all operations
- ✅ **Flexible Queries**: Pagination, sorting, filtering, and multiple response formats
- ✅ **Schema Fetching**: Dynamic schema fetching with localization support
- ✅ **Type Safety**: Item type and creator type constants for IDE autocomplete
- ✅ **CLI Tool**: Command-line interface with environment variable support
- ✅ **Comprehensive Testing**: Unit tests with mock servers and integration tests for live/local APIs

## Installation

```bash
go get github.com/Epistemic-Technology/zotero
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/Epistemic-Technology/zotero/zotero"
)

func main() {
    ctx := context.Background()
    
    // Create a client
    client, err := zotero.NewClient(
        "12345",                    // Library ID
        zotero.LibraryTypeUser,     // Library type (user or group)
        zotero.WithAPIKey("your-api-key"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Fetch items
    items, err := client.Items(ctx, &zotero.QueryParams{
        Limit: 10,
        ItemType: []string{zotero.ItemTypeJournalArticle},
    })
    if err != nil {
        log.Fatal(err)
    }
    
    for _, item := range items {
        fmt.Printf("%s: %s\n", item.Key, item.Data.Title)
    }
}
```

### Creating Items

```go
// Get a template for the item type
template, err := client.NewItemTemplate(ctx, zotero.ItemTypeBook)
if err != nil {
    log.Fatal(err)
}

// Create a new book
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

resp, err := client.CreateItems(ctx, []zotero.Item{item})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created %d items\n", len(resp.Success))
```

### File Operations

```go
// Upload an attachment
attachment, err := client.UploadAttachment(ctx, parentItemKey, "/path/to/file.pdf", "file.pdf", "application/pdf")
if err != nil {
    log.Fatal(err)
}

// Download an attachment
fullPath, err := client.Dump(ctx, "ABCD1234", "", "/path/to/downloads")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("File saved to: %s\n", fullPath)
```

## CLI Tool

The project includes a command-line tool for interacting with the Zotero API:

```bash
# Build the CLI
make zotero-cli

# Set environment variables (recommended)
export ZOTERO_API_KEY=your_key
export ZOTERO_LIBRARY_ID=your_library_id
export ZOTERO_LIBRARY_TYPE=user

# Use the CLI
bin/zotero-cli items -limit 10
bin/zotero-cli items -itemtype journalArticle -limit 10
bin/zotero-cli collections
bin/zotero-cli download -item ABC123 -path ./downloads
```

## Development

### Testing

```bash
# Run unit tests (fast, no credentials required)
make test-unit

# Run integration tests (requires .env with API credentials)
make test-integration

# Run all tests
make test-all
```

See [tests/README.md](tests/README.md) for detailed testing documentation.

### Building

```bash
make build              # Build all binaries
make zotero-cli         # Build only the CLI tool
make help               # Show all available targets
```

## Documentation

For comprehensive documentation including:
- Detailed API usage examples
- Write operations (create, update, delete)
- Batch operations
- Query parameters and filtering
- Schema fetching
- Testing strategies

See [CLAUDE.md](CLAUDE.md) for complete project documentation.

## Credit

This library is heavily inspired by the [Pyzotero](https://github.com/urschrei/pyzotero) library and can be largely considered a port of it to the Go programming language. Pyzotero uses the [Blue Oak Model License](https://github.com/urschrei/pyzotero/blob/main/LICENSE.md).

## References

- [Zotero Web API v3 Documentation](https://www.zotero.org/support/dev/web_api/v3/start)
- [Pyzotero documentation](https://pyzotero.readthedocs.io/en/latest/) - Python implementation serving as a reference for this library

## License

[Add your license information here]
