# Integration Tests

This directory contains integration tests for the Zotero Go client library. These tests interact with real Zotero APIs (either the live Zotero Web API or a local REST API) and require valid credentials.

## Quick Start

### Option 1: Using `.env` file (Recommended for local development)

1. **Copy the environment template:**
   ```bash
   cp .env.example .env
   ```

2. **Add your credentials to `.env`:**
   ```bash
   ZOTERO_API_KEY=your_api_key_here
   ZOTERO_LIBRARY_ID=your_library_id_here
   ZOTERO_LIBRARY_TYPE=user
   TEST_API_URL=https://api.zotero.org
   ```

3. **Run the integration tests:**
   ```bash
   make test-integration
   # or
   go test ./tests -v
   ```

### Option 2: Using shell environment variables (Recommended for permanent setup)

**Benefit:** These same environment variables work for both integration tests AND the CLI tool!

1. **Add to your shell profile** (e.g., `~/.zshrc`, `~/.bashrc`):
   ```bash
   export ZOTERO_API_KEY=your_api_key_here
   export ZOTERO_LIBRARY_ID=your_library_id_here
   export ZOTERO_LIBRARY_TYPE=user
   export TEST_API_URL=https://api.zotero.org  # for integration tests
   ```

2. **Reload your shell:**
   ```bash
   source ~/.zshrc  # or ~/.bashrc
   ```

3. **Run the integration tests:**
   ```bash
   make test-integration
   # or
   go test ./tests -v
   ```

4. **Use the CLI without flags:**
   ```bash
   bin/zotero-cli items -limit 10
   bin/zotero-cli collections
   ```

**Note:** Environment variables set in your shell take precedence. If both exist, `.env` values will only be used if the shell environment variable is not set.

## Getting Credentials

### API Key

1. Log in to your Zotero account at https://www.zotero.org
2. Go to Settings → Feeds/API → Create new private key
3. Set the permissions you need (read-only is sufficient for these tests)
4. Copy the generated API key

### Library ID

Your library ID can be found:
- On the API key creation page (shown as "Your userID for use in API calls")
- In the URL when viewing your library: `https://www.zotero.org/username` (the numeric ID is shown in some API responses)

For **group libraries**: Use the group ID instead of your user ID

### Library Type

- `user` - For your personal library (default)
- `group` - For group libraries (requires group ID as LIBRARY_ID)

## Testing Against Different APIs

### Live Zotero Web API (Default)

```bash
TEST_API_URL=https://api.zotero.org
```

This is the default if `TEST_API_URL` is not set.

### Local Zotero REST API

If you're running Zotero desktop application with the local REST API enabled:

```bash
TEST_API_URL=http://localhost:23119
```

The local API typically runs on port 23119. You'll need to:
1. Install Zotero desktop application
2. Enable the local API in Zotero preferences
3. Use your local library ID

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ZOTERO_API_KEY` | Yes | - | Your Zotero API key for authentication |
| `ZOTERO_LIBRARY_ID` | Yes | - | Your user ID or group ID |
| `ZOTERO_LIBRARY_TYPE` | No | `user` | Library type: `user` or `group` |
| `TEST_API_URL` | No | `https://api.zotero.org` | API endpoint URL |

## Running Tests

### Run all integration tests
```bash
make test-integration
# or
go test ./tests -v
```

### Run a specific test
```bash
go test ./tests -v -run TestItems
go test ./tests -v -run TestCollections
```

### Run tests with more verbose output
```bash
go test ./tests -v -count=1
```

### Skip integration tests (automatic)
If `ZOTERO_API_KEY` or `ZOTERO_LIBRARY_ID` are not set, tests will automatically skip with a message:
```
--- SKIP: TestItems (0.00s)
    integration_test.go:XX: Skipping integration test: ZOTERO_API_KEY and ZOTERO_LIBRARY_ID not set
```

## What's Tested

The integration tests cover:

### Item Operations
- `Items()` - List all items with pagination
- `Top()` - Get top-level items (no parents)
- `Item()` - Get a specific item by key
- `Children()` - Get child items (attachments, notes)
- `Trash()` - Get items in trash

### Collection Operations
- `Collections()` - List all collections
- `CollectionsTop()` - Get top-level collections
- `Collection()` - Get a specific collection
- `CollectionItems()` - Get items in a collection
- `CollectionItemsTop()` - Get top-level items in a collection

### Tag Operations
- `Tags()` - List all tags
- `ItemTags()` - Get tags associated with items
- `CollectionTags()` - Get tags associated with collections

### Group Operations
- `Groups()` - List user's groups (user libraries only)

### Utility Operations
- `NumItems()` - Get total item count
- `LastModifiedVersion()` - Get library version
- `Deleted()` - Get deleted items since a version

### Query Features
- Pagination (limit, start)
- Sorting (title, dateAdded, dateModified)
- Quick search
- Filtering

## Test Data Requirements

Some tests require specific library contents:
- **TestChildren** - Requires at least one item with attachments or notes
- **TestCollectionItems** - Requires at least one collection with items
- **TestGroups** - Requires a user library (not group library)

Tests will automatically skip if the required data is not found in your library.

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Run unit tests
        run: make test-unit
      
      - name: Run integration tests
        if: ${{ secrets.ZOTERO_API_KEY != '' }}
        env:
          ZOTERO_API_KEY: ${{ secrets.ZOTERO_API_KEY }}
          ZOTERO_LIBRARY_ID: ${{ secrets.ZOTERO_LIBRARY_ID }}
          ZOTERO_LIBRARY_TYPE: user
        run: make test-integration
```

Store `ZOTERO_API_KEY` and `ZOTERO_LIBRARY_ID` as repository secrets.

## Troubleshooting

### Tests are skipped
- Verify `.env` file exists and has correct values
- Check that environment variables are exported: `echo $ZOTERO_API_KEY`
- Try running with explicit env vars: `ZOTERO_API_KEY=xxx ZOTERO_LIBRARY_ID=yyy go test ./tests -v`

### Authentication errors
- Verify your API key is valid and not expired
- Check that the API key has the necessary permissions (read access)
- Ensure you're using the correct library ID for your account

### Connection errors with local API
- Verify Zotero desktop is running
- Check that the local API is enabled in Zotero preferences
- Verify the port (default: 23119) matches your configuration
- Try accessing http://localhost:23119 in your browser

### Rate limiting
- The live API has rate limits (default: 1 request/second)
- Tests disable rate limiting by default, but the server may still enforce limits
- If you hit rate limits, wait a few minutes before retrying

## Best Practices

1. **Use a test library**: Consider creating a separate Zotero library for testing
2. **Read-only API keys**: Create API keys with read-only permissions for safety
3. **Don't commit credentials**: The `.env` file is in `.gitignore` - never commit it
4. **Local API for development**: Use the local REST API for faster development/testing
5. **CI/CD secrets**: Store credentials as encrypted secrets in your CI/CD system

## Contributing

When adding new integration tests:
1. Use `skipIfNoCredentials(t)` to handle missing credentials gracefully
2. Add appropriate test data requirements to this README
3. Verify tests work against both live and local APIs
4. Keep tests idempotent (don't modify library state)
5. Use descriptive test names and log useful information
