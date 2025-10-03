package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

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
		itemsCmd.Parse(os.Args[2:])

		if libraryID == "" {
			fmt.Println("Error: -library is required")
			itemsCmd.PrintDefaults()
			os.Exit(1)
		}

		listItems(libraryID, libraryType, apiKey, verbose, *limit, *start)

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

func listItems(libraryID, libraryType, apiKey string, verbose bool, limit, start int) {
	client := createClient(libraryID, libraryType, apiKey, verbose)
	ctx := context.Background()

	params := &zotero.QueryParams{
		Limit: limit,
		Start: start,
	}

	items, err := client.Items(ctx, params)
	if err != nil {
		fmt.Printf("Error fetching items: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Retrieved %d items:\n\n", len(items))
	printJSON(items)
}

func getItem(libraryID, libraryType, apiKey string, verbose bool, itemKey string) {
	client := createClient(libraryID, libraryType, apiKey, verbose)
	ctx := context.Background()

	item, err := client.Item(ctx, itemKey, nil)
	if err != nil {
		fmt.Printf("Error fetching item: %v\n", err)
		os.Exit(1)
	}

	printJSON(item)
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
	printJSON(collections)
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
	printJSON(groups)
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
