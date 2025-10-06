package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Epistemic-Technology/zotero/zotero"
)

func main() {
	// Define subcommands
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Common flags
	var (
		apiKey      string
		libraryID   string
		libraryType string
		verbose     bool
	)

	// Get values from environment if not provided
	envAPIKey := os.Getenv("ZOTERO_API_KEY")
	envLibraryID := os.Getenv("ZOTERO_LIBRARY_ID")
	envLibraryType := os.Getenv("ZOTERO_LIBRARY_TYPE")
	if envLibraryType == "" {
		envLibraryType = "user"
	}

	switch os.Args[1] {
	case "items":
		itemsCmd := flag.NewFlagSet("items", flag.ExitOnError)
		itemsCmd.StringVar(&apiKey, "key", envAPIKey, "Zotero API key (or set ZOTERO_API_KEY)")
		itemsCmd.StringVar(&libraryID, "library", envLibraryID, "Library ID (or set ZOTERO_LIBRARY_ID)")
		itemsCmd.StringVar(&libraryType, "type", envLibraryType, "Library type: user or group (or set ZOTERO_LIBRARY_TYPE)")
		itemsCmd.BoolVar(&verbose, "v", false, "Enable verbose logging")
		limit := itemsCmd.Int("limit", 25, "Number of items to retrieve")
		start := itemsCmd.Int("start", 0, "Starting index")
		itemType := itemsCmd.String("itemtype", "", "Filter by item type(s), comma-separated; prefix with '-' to exclude (e.g., 'journalArticle' or '-annotation')")
		itemsCmd.Parse(os.Args[2:])

		if libraryID == "" {
			fmt.Println("Error: -library is required")
			itemsCmd.PrintDefaults()
			os.Exit(1)
		}

		listItems(libraryID, libraryType, apiKey, verbose, *limit, *start, *itemType)

	case "item":
		itemCmd := flag.NewFlagSet("item", flag.ExitOnError)
		itemCmd.StringVar(&apiKey, "key", envAPIKey, "Zotero API key (or set ZOTERO_API_KEY)")
		itemCmd.StringVar(&libraryID, "library", envLibraryID, "Library ID (or set ZOTERO_LIBRARY_ID)")
		itemCmd.StringVar(&libraryType, "type", envLibraryType, "Library type: user or group (or set ZOTERO_LIBRARY_TYPE)")
		itemCmd.BoolVar(&verbose, "v", false, "Enable verbose logging")
		itemKey := itemCmd.String("item", "", "Item key (required)")
		itemCmd.Parse(os.Args[2:])

		if libraryID == "" || *itemKey == "" {
			fmt.Println("Error: -library and -item are required")
			itemCmd.PrintDefaults()
			os.Exit(1)
		}

		getItem(libraryID, libraryType, apiKey, verbose, *itemKey)

	case "collections":
		collectionsCmd := flag.NewFlagSet("collections", flag.ExitOnError)
		collectionsCmd.StringVar(&apiKey, "key", envAPIKey, "Zotero API key (or set ZOTERO_API_KEY)")
		collectionsCmd.StringVar(&libraryID, "library", envLibraryID, "Library ID (or set ZOTERO_LIBRARY_ID)")
		collectionsCmd.StringVar(&libraryType, "type", envLibraryType, "Library type: user or group (or set ZOTERO_LIBRARY_TYPE)")
		collectionsCmd.BoolVar(&verbose, "v", false, "Enable verbose logging")
		collectionsCmd.Parse(os.Args[2:])

		if libraryID == "" {
			fmt.Println("Error: -library is required")
			collectionsCmd.PrintDefaults()
			os.Exit(1)
		}

		listCollections(libraryID, libraryType, apiKey, verbose)

	case "groups":
		groupsCmd := flag.NewFlagSet("groups", flag.ExitOnError)
		groupsCmd.StringVar(&apiKey, "key", envAPIKey, "Zotero API key (or set ZOTERO_API_KEY)")
		groupsCmd.BoolVar(&verbose, "v", false, "Enable verbose logging")
		userID := groupsCmd.String("user", "", "User ID (required for groups)")
		groupsCmd.Parse(os.Args[2:])

		if *userID == "" {
			fmt.Println("Error: -user is required")
			groupsCmd.PrintDefaults()
			os.Exit(1)
		}

		listGroups(*userID, apiKey, verbose)

	case "create":
		createCmd := flag.NewFlagSet("create", flag.ExitOnError)
		createCmd.StringVar(&apiKey, "key", envAPIKey, "Zotero API key (or set ZOTERO_API_KEY)")
		createCmd.StringVar(&libraryID, "library", envLibraryID, "Library ID (or set ZOTERO_LIBRARY_ID)")
		createCmd.StringVar(&libraryType, "type", envLibraryType, "Library type: user or group (or set ZOTERO_LIBRARY_TYPE)")
		createCmd.BoolVar(&verbose, "v", false, "Enable verbose logging")
		itemType := createCmd.String("itemtype", zotero.ItemTypeJournalArticle, "Item type (e.g., book, journalArticle, webpage)")
		title := createCmd.String("title", "", "Item title (required)")
		authors := createCmd.String("authors", "", "Authors (comma-separated, format: 'First Last, First Last')")
		file := createCmd.String("file", "", "Optional: Path to file to attach to the item")
		contentType := createCmd.String("contenttype", "application/pdf", "MIME type of the file (used with -file)")
		createCmd.Parse(os.Args[2:])

		if libraryID == "" || *title == "" {
			fmt.Println("Error: -library and -title are required")
			createCmd.PrintDefaults()
			os.Exit(1)
		}

		if apiKey == "" {
			fmt.Println("Error: API key required for write operations")
			createCmd.PrintDefaults()
			os.Exit(1)
		}

		createItem(libraryID, libraryType, apiKey, verbose, *itemType, *title, *authors, *file, *contentType)

	case "upload":
		uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
		uploadCmd.StringVar(&apiKey, "key", envAPIKey, "Zotero API key (or set ZOTERO_API_KEY)")
		uploadCmd.StringVar(&libraryID, "library", envLibraryID, "Library ID (or set ZOTERO_LIBRARY_ID)")
		uploadCmd.StringVar(&libraryType, "type", envLibraryType, "Library type: user or group (or set ZOTERO_LIBRARY_TYPE)")
		uploadCmd.BoolVar(&verbose, "v", false, "Enable verbose logging")
		file := uploadCmd.String("file", "", "Path to file to upload (required)")
		parentItem := uploadCmd.String("parent", "", "Parent item key (empty for standalone attachment)")
		contentType := uploadCmd.String("contenttype", "application/pdf", "MIME type of the file")
		uploadCmd.Parse(os.Args[2:])

		if libraryID == "" || *file == "" {
			fmt.Println("Error: -library and -file are required")
			uploadCmd.PrintDefaults()
			os.Exit(1)
		}

		if apiKey == "" {
			fmt.Println("Error: API key required for write operations")
			uploadCmd.PrintDefaults()
			os.Exit(1)
		}

		uploadFile(libraryID, libraryType, apiKey, verbose, *file, *parentItem, *contentType)

	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Zotero CLI - Interact with the Zotero Web API")
	fmt.Println("\nUsage:")
	fmt.Println("  zotero-cli <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  items         List items in a library")
	fmt.Println("  item          Get a specific item")
	fmt.Println("  collections   List collections in a library")
	fmt.Println("  groups        List groups for a user")
	fmt.Println("  create        Create a new item")
	fmt.Println("  upload        Upload a file attachment")
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  ZOTERO_API_KEY       API key for authentication")
	fmt.Println("  ZOTERO_LIBRARY_ID    Library ID (default for commands)")
	fmt.Println("  ZOTERO_LIBRARY_TYPE  Library type: user or group (default: user)")
	fmt.Println("\nExamples:")
	fmt.Println("  zotero-cli items -library 12345 -type user -limit 10")
	fmt.Println("  zotero-cli item -library 12345 -item ABC123")
	fmt.Println("  zotero-cli collections -library 12345")
	fmt.Println("  zotero-cli groups -user 12345")
	fmt.Println("  zotero-cli create -title 'My Paper' -authors 'John Doe, Jane Smith'")
	fmt.Println("  zotero-cli create -title 'Research Article' -file paper.pdf")
	fmt.Println("  zotero-cli upload -file paper.pdf -parent ABC123")
}

func listItems(libraryID, libraryType, apiKey string, verbose bool, limit, start int, itemType string) {
	client := createClient(libraryID, libraryType, apiKey, verbose)
	ctx := context.Background()

	params := &zotero.QueryParams{
		Limit: limit,
		Start: start,
	}

	// Parse itemType filter if provided
	if itemType != "" {
		itemTypes := strings.Split(itemType, ",")
		for i, it := range itemTypes {
			itemTypes[i] = strings.TrimSpace(it)
		}
		params.ItemType = itemTypes
	}

	items, err := client.Items(ctx, params)
	if err != nil {
		fmt.Printf("Error fetching items: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Retrieved %d items:\n\n", len(items))
	printItemsTable(items)
}

func getItem(libraryID, libraryType, apiKey string, verbose bool, itemKey string) {
	client := createClient(libraryID, libraryType, apiKey, verbose)
	ctx := context.Background()

	item, err := client.Item(ctx, itemKey, nil)
	if err != nil {
		fmt.Printf("Error fetching item: %v\n", err)
		os.Exit(1)
	}

	printItemDetails(item)
}

func listCollections(libraryID, libraryType, apiKey string, verbose bool) {
	client := createClient(libraryID, libraryType, apiKey, verbose)
	ctx := context.Background()

	collections, err := client.Collections(ctx, nil)
	if err != nil {
		fmt.Printf("Error fetching collections: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Retrieved %d collections:\n\n", len(collections))
	printCollectionsTable(collections)
}

func listGroups(userID, apiKey string, verbose bool) {
	client := createClient(userID, string(zotero.LibraryTypeUser), apiKey, verbose)
	ctx := context.Background()

	groups, err := client.Groups(ctx, nil)
	if err != nil {
		fmt.Printf("Error fetching groups: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Retrieved %d groups:\n\n", len(groups))
	printGroupsTable(groups)
}

func createClient(libraryID, libraryType, apiKey string, verbose bool) *zotero.Client {
	libType := zotero.LibraryTypeUser
	if libraryType == "group" {
		libType = zotero.LibraryTypeGroup
	}

	opts := []zotero.ClientOption{}
	if apiKey != "" {
		opts = append(opts, zotero.WithAPIKey(apiKey))
	}

	if verbose {
		logger := log.New(os.Stderr, "[zotero] ", log.LstdFlags)
		opts = append(opts, zotero.WithLogger(logger))
	}

	return zotero.NewClient(libraryID, libType, opts...)
}

func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// printItemsTable displays items in a formatted table
func printItemsTable(items []zotero.Item) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\tTYPE\tTITLE\tCREATORS\tDATE")
	fmt.Fprintln(w, "---\t----\t-----\t--------\t----")

	for _, item := range items {
		key := item.Key
		itemType := item.Data.ItemType
		title := truncate(item.Data.Title, 40)
		creators := formatCreators(item.Data.Creators)
		date := item.Data.DateAdded
		if len(date) > 10 {
			date = date[:10] // Show only date part (YYYY-MM-DD)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", key, itemType, title, creators, date)
	}
	w.Flush()
}

// printItemDetails displays a single item with detailed field information
func printItemDetails(item *zotero.Item) {
	fmt.Printf("Item: %s\n", item.Key)
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("Type:     %s\n", item.Data.ItemType)

	if item.Data.Title != "" {
		fmt.Printf("Title:    %s\n", item.Data.Title)
	}

	if len(item.Data.Creators) > 0 {
		fmt.Println("Creators:")
		for _, creator := range item.Data.Creators {
			name := creator.Name
			if name == "" {
				name = fmt.Sprintf("%s %s", creator.FirstName, creator.LastName)
			}
			fmt.Printf("  - %s (%s)\n", name, creator.CreatorType)
		}
	}

	if item.Data.AbstractNote != "" {
		fmt.Printf("Abstract: %s\n", truncate(item.Data.AbstractNote, 200))
	}

	if item.Data.DateAdded != "" {
		fmt.Printf("Added:    %s\n", item.Data.DateAdded)
	}

	if item.Data.DateModified != "" {
		fmt.Printf("Modified: %s\n", item.Data.DateModified)
	}

	if len(item.Data.Tags) > 0 {
		tags := make([]string, len(item.Data.Tags))
		for i, tag := range item.Data.Tags {
			tags[i] = tag.Tag
		}
		fmt.Printf("Tags:     %s\n", strings.Join(tags, ", "))
	}

	if len(item.Data.Collections) > 0 {
		fmt.Printf("Collections: %d\n", len(item.Data.Collections))
	}

	if item.Meta.NumChildren > 0 {
		fmt.Printf("Children: %d\n", item.Meta.NumChildren)
	}

	fmt.Printf("Version:  %d\n", item.Version)
}

// printCollectionsTable displays collections in a formatted table
func printCollectionsTable(collections []zotero.Collection) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\tNAME\tPARENT\tITEMS")
	fmt.Fprintln(w, "---\t----\t------\t-----")

	for _, coll := range collections {
		key := coll.Key
		name := truncate(coll.Data.Name, 40)
		parent := string(coll.Data.ParentCollection)
		if parent == "" {
			parent = "-"
		} else {
			parent = truncate(parent, 8)
		}
		items := fmt.Sprintf("%d", coll.Meta.NumItems)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", key, name, parent, items)
	}
	w.Flush()
}

// printGroupsTable displays groups in a formatted table
func printGroupsTable(groups []zotero.Group) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTYPE\tITEMS\tMEMBERS")
	fmt.Fprintln(w, "--\t----\t----\t-----\t-------")

	for _, group := range groups {
		id := fmt.Sprintf("%d", group.ID)
		name := truncate(group.Name, 30)
		groupType := group.Type
		items := fmt.Sprintf("%d", group.Meta.NumItems)
		members := fmt.Sprintf("%d", len(group.Members))

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, name, groupType, items, members)
	}
	w.Flush()
}

// formatCreators formats a list of creators for display
func formatCreators(creators []zotero.Creator) string {
	if len(creators) == 0 {
		return "-"
	}

	names := make([]string, 0, len(creators))
	for i, creator := range creators {
		if i >= 2 {
			names = append(names, fmt.Sprintf("(+%d more)", len(creators)-2))
			break
		}

		name := creator.Name
		if name == "" {
			if creator.LastName != "" {
				name = creator.LastName
			} else {
				name = creator.FirstName
			}
		}
		names = append(names, name)
	}

	return truncate(strings.Join(names, ", "), 30)
}

// truncate truncates a string to a maximum length with ellipsis
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// createItem creates a new item in the library
func createItem(libraryID, libraryType, apiKey string, verbose bool, itemType, title, authors, file, contentType string) {
	client := createClient(libraryID, libraryType, apiKey, verbose)
	ctx := context.Background()

	// Parse authors
	var creators []zotero.Creator
	if authors != "" {
		authorList := strings.Split(authors, ",")
		for _, author := range authorList {
			author = strings.TrimSpace(author)
			parts := strings.Fields(author)
			if len(parts) >= 2 {
				creators = append(creators, zotero.Creator{
					CreatorType: zotero.CreatorTypeAuthor,
					FirstName:   strings.Join(parts[:len(parts)-1], " "),
					LastName:    parts[len(parts)-1],
				})
			} else if len(parts) == 1 {
				creators = append(creators, zotero.Creator{
					CreatorType: zotero.CreatorTypeAuthor,
					Name:        parts[0],
				})
			}
		}
	}

	// Create the item
	item := zotero.Item{
		Data: zotero.ItemData{
			ItemType: itemType,
			Title:    title,
			Creators: creators,
		},
	}

	resp, err := client.CreateItems(ctx, []zotero.Item{item})
	if err != nil {
		fmt.Printf("Error creating item: %v\n", err)
		os.Exit(1)
	}

	var itemKey string
	if len(resp.Success) > 0 {
		for idx, key := range resp.Success {
			if keyStr, ok := key.(string); ok {
				itemKey = keyStr
				fmt.Printf("Successfully created item with key: %s (index: %s)\n", keyStr, idx)
			}
		}
	}

	if len(resp.Failed) > 0 {
		fmt.Println("\nFailed items:")
		for idx, failure := range resp.Failed {
			fmt.Printf("  Index %s: %d - %s\n", idx, failure.Code, failure.Message)
		}
		os.Exit(1)
	}

	// Upload file attachment if specified
	if file != "" && itemKey != "" {
		fmt.Printf("\nUploading attachment: %s\n", file)
		attachment, err := client.UploadAttachment(ctx, itemKey, file, "", contentType)
		if err != nil {
			fmt.Printf("Error uploading attachment: %v\n", err)
			fmt.Println("Note: Item was created successfully, but attachment upload failed")
			os.Exit(1)
		}
		fmt.Printf("Successfully attached file!\n")
		fmt.Printf("Attachment Key: %s\n", attachment.Key)
		fmt.Printf("Filename: %s\n", attachment.Data.Filename)
	}
}

// uploadFile uploads a file as an attachment
func uploadFile(libraryID, libraryType, apiKey string, verbose bool, filepath, parentItem, contentType string) {
	client := createClient(libraryID, libraryType, apiKey, verbose)
	ctx := context.Background()

	fmt.Printf("Uploading file: %s\n", filepath)
	if parentItem != "" {
		fmt.Printf("Parent item: %s\n", parentItem)
	} else {
		fmt.Println("Creating standalone attachment")
	}

	item, err := client.UploadAttachment(ctx, parentItem, filepath, "", contentType)
	if err != nil {
		fmt.Printf("Error uploading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSuccessfully uploaded attachment!\n")
	fmt.Printf("Key: %s\n", item.Key)
	fmt.Printf("Title: %s\n", item.Data.Title)
	fmt.Printf("Content Type: %s\n", item.Data.ContentType)
	fmt.Printf("Filename: %s\n", item.Data.Filename)
}
