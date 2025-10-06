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
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  ZOTERO_API_KEY       API key for authentication")
	fmt.Println("  ZOTERO_LIBRARY_ID    Library ID (default for commands)")
	fmt.Println("  ZOTERO_LIBRARY_TYPE  Library type: user or group (default: user)")
	fmt.Println("\nExamples:")
	fmt.Println("  zotero-cli items -library 12345 -type user -limit 10")
	fmt.Println("  zotero-cli item -library 12345 -item ABC123")
	fmt.Println("  zotero-cli collections -library 12345")
	fmt.Println("  zotero-cli groups -user 12345")
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
		parent := coll.Data.ParentCollection
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
