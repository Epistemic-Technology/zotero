package zotero

import "time"

// Item represents a Zotero item (book, article, note, etc.)
type Item struct {
	// Core metadata
	Key     string  `json:"key,omitempty"`
	Version int     `json:"version,omitempty"`
	Library Library `json:"library,omitempty"`
	Links   Links   `json:"links,omitempty"`
	Meta    Meta    `json:"meta,omitempty"`

	// Item data
	Data ItemData `json:"data,omitempty"`
}

// ItemData contains the actual item content
type ItemData struct {
	Key          string    `json:"key,omitempty"`
	Version      int       `json:"version,omitempty"`
	ItemType     string    `json:"itemType"`
	Title        string    `json:"title,omitempty"`
	Creators     []Creator `json:"creators,omitempty"`
	AbstractNote string    `json:"abstractNote,omitempty"`
	Tags         []Tag     `json:"tags,omitempty"`
	Collections  []string  `json:"collections,omitempty"`
	Relations    Relations `json:"relations,omitempty"`
	DateAdded    string    `json:"dateAdded,omitempty"`
	DateModified string    `json:"dateModified,omitempty"`

	// Additional fields that vary by item type
	Extra map[string]any `json:"-"`
}

// Creator represents a creator (author, editor, etc.)
type Creator struct {
	CreatorType string `json:"creatorType"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Name        string `json:"name,omitempty"` // For single-field mode
}

// Tag represents an item tag
type Tag struct {
	Tag  string `json:"tag"`
	Type int    `json:"type,omitempty"` // 0 for automatic, 1 for manual
}

// Relations represents relationships to other items
type Relations struct {
	OwlSameAs      any `json:"owl:sameAs,omitempty"`
	DCRelation     any `json:"dc:relation,omitempty"`
	DCReplaces     any `json:"dc:replaces,omitempty"`
	DCIsReplacedBy any `json:"dc:isReplacedBy,omitempty"`
}

// Collection represents a Zotero collection
type Collection struct {
	// Core metadata
	Key     string  `json:"key,omitempty"`
	Version int     `json:"version,omitempty"`
	Library Library `json:"library,omitempty"`
	Links   Links   `json:"links,omitempty"`
	Meta    Meta    `json:"meta,omitempty"`

	// Collection data
	Data CollectionData `json:"data,omitempty"`
}

// CollectionData contains the actual collection content
type CollectionData struct {
	Key              string    `json:"key,omitempty"`
	Version          int       `json:"version,omitempty"`
	Name             string    `json:"name"`
	ParentCollection string    `json:"parentCollection,omitempty"`
	Relations        Relations `json:"relations,omitempty"`
}

// Search represents a saved search
type Search struct {
	// Core metadata
	Key     string  `json:"key,omitempty"`
	Version int     `json:"version,omitempty"`
	Library Library `json:"library,omitempty"`
	Links   Links   `json:"links,omitempty"`
	Meta    Meta    `json:"meta,omitempty"`

	// Search data
	Data SearchData `json:"data,omitempty"`
}

// SearchData contains the actual search content
type SearchData struct {
	Key        string            `json:"key,omitempty"`
	Version    int               `json:"version,omitempty"`
	Name       string            `json:"name"`
	Conditions []SearchCondition `json:"conditions"`
}

// SearchCondition represents a single search condition
type SearchCondition struct {
	Condition string `json:"condition"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`
}

// Group represents a Zotero group
type Group struct {
	ID          int       `json:"id"`
	Version     int       `json:"version"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // "Private" or "Public"
	Description string    `json:"description,omitempty"`
	URL         string    `json:"url,omitempty"`
	LibraryID   int       `json:"libraryID,omitempty"`
	Owner       int       `json:"owner"`
	Members     []int     `json:"members,omitempty"`
	Admins      []int     `json:"admins,omitempty"`
	FileEditing string    `json:"fileEditing,omitempty"`
	Meta        GroupMeta `json:"meta,omitempty"`
}

// GroupMeta contains group metadata
type GroupMeta struct {
	Created      time.Time `json:"created,omitempty"`
	LastModified time.Time `json:"lastModified,omitempty"`
	NumItems     int       `json:"numItems,omitempty"`
}

// Library represents library information
type Library struct {
	Type  string `json:"type"` // "user" or "group"
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Links Links  `json:"links,omitempty"`
}

// Links contains hypermedia links
type Links struct {
	Self      Link `json:"self,omitempty"`
	Alternate Link `json:"alternate,omitempty"`
	Up        Link `json:"up,omitempty"`
	Enclosure Link `json:"enclosure,omitempty"`
}

// Link represents a single hypermedia link
type Link struct {
	Href string `json:"href"`
	Type string `json:"type,omitempty"`
}

// Meta contains response metadata
type Meta struct {
	CreatorSummary string `json:"creatorSummary,omitempty"`
	ParsedDate     string `json:"parsedDate,omitempty"`
	NumChildren    int    `json:"numChildren,omitempty"`
	NumCollections int    `json:"numCollections,omitempty"`
	NumItems       int    `json:"numItems,omitempty"`
}

// ItemType represents an item type definition
type ItemType struct {
	ItemType      string `json:"itemType"`
	LocalizedName string `json:"localized,omitempty"`
}

// ItemField represents a field definition
type ItemField struct {
	Field         string `json:"field"`
	LocalizedName string `json:"localized,omitempty"`
}

// CreatorType represents a creator type definition
type CreatorType struct {
	CreatorType   string `json:"creatorType"`
	LocalizedName string `json:"localized,omitempty"`
}

// WriteResponse represents the response from write operations
type WriteResponse struct {
	Success   map[string]any         `json:"success,omitempty"`
	Unchanged map[string]any         `json:"unchanged,omitempty"`
	Failed    map[string]FailedWrite `json:"failed,omitempty"`
}

// FailedWrite represents a failed write operation
type FailedWrite struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// DeletedContent represents deleted items/collections
type DeletedContent struct {
	Items       []string `json:"items,omitempty"`
	Collections []string `json:"collections,omitempty"`
	Searches    []string `json:"searches,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}
