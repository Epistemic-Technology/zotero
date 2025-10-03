package zotero

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestItemUnmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/item.json")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	var item Item
	if err := json.Unmarshal(data, &item); err != nil {
		t.Fatalf("failed to unmarshal item: %v", err)
	}

	// Test core metadata
	if item.Key != "ABCD1234" {
		t.Errorf("Key = %v, want ABCD1234", item.Key)
	}
	if item.Version != 100 {
		t.Errorf("Version = %v, want 100", item.Version)
	}

	// Test library
	if item.Library.Type != "user" {
		t.Errorf("Library.Type = %v, want user", item.Library.Type)
	}
	if item.Library.ID != 12345 {
		t.Errorf("Library.ID = %v, want 12345", item.Library.ID)
	}
	if item.Library.Name != "testuser" {
		t.Errorf("Library.Name = %v, want testuser", item.Library.Name)
	}

	// Test links
	if item.Links.Self.Href != "https://api.zotero.org/users/12345/items/ABCD1234" {
		t.Errorf("Links.Self.Href = %v, want https://api.zotero.org/users/12345/items/ABCD1234", item.Links.Self.Href)
	}

	// Test meta
	if item.Meta.CreatorSummary != "Doe" {
		t.Errorf("Meta.CreatorSummary = %v, want Doe", item.Meta.CreatorSummary)
	}
	if item.Meta.NumChildren != 2 {
		t.Errorf("Meta.NumChildren = %v, want 2", item.Meta.NumChildren)
	}

	// Test data
	if item.Data.ItemType != "book" {
		t.Errorf("Data.ItemType = %v, want book", item.Data.ItemType)
	}
	if item.Data.Title != "Test Book Title" {
		t.Errorf("Data.Title = %v, want Test Book Title", item.Data.Title)
	}
	if item.Data.AbstractNote != "This is a test abstract" {
		t.Errorf("Data.AbstractNote = %v, want This is a test abstract", item.Data.AbstractNote)
	}

	// Test creators
	if len(item.Data.Creators) != 1 {
		t.Fatalf("len(Creators) = %v, want 1", len(item.Data.Creators))
	}
	if item.Data.Creators[0].CreatorType != "author" {
		t.Errorf("Creators[0].CreatorType = %v, want author", item.Data.Creators[0].CreatorType)
	}
	if item.Data.Creators[0].FirstName != "John" {
		t.Errorf("Creators[0].FirstName = %v, want John", item.Data.Creators[0].FirstName)
	}
	if item.Data.Creators[0].LastName != "Doe" {
		t.Errorf("Creators[0].LastName = %v, want Doe", item.Data.Creators[0].LastName)
	}

	// Test tags
	if len(item.Data.Tags) != 2 {
		t.Fatalf("len(Tags) = %v, want 2", len(item.Data.Tags))
	}
	if item.Data.Tags[0].Tag != "test" {
		t.Errorf("Tags[0].Tag = %v, want test", item.Data.Tags[0].Tag)
	}
	if item.Data.Tags[1].Tag != "example" {
		t.Errorf("Tags[1].Tag = %v, want example", item.Data.Tags[1].Tag)
	}
	if item.Data.Tags[1].Type != 1 {
		t.Errorf("Tags[1].Type = %v, want 1", item.Data.Tags[1].Type)
	}

	// Test collections
	if len(item.Data.Collections) != 1 {
		t.Fatalf("len(Collections) = %v, want 1", len(item.Data.Collections))
	}
	if item.Data.Collections[0] != "COLL1234" {
		t.Errorf("Collections[0] = %v, want COLL1234", item.Data.Collections[0])
	}
}

func TestItemsUnmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/items.json")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	var items []Item
	if err := json.Unmarshal(data, &items); err != nil {
		t.Fatalf("failed to unmarshal items: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("len(items) = %v, want 2", len(items))
	}

	// Test first item
	if items[0].Key != "ABCD1234" {
		t.Errorf("items[0].Key = %v, want ABCD1234", items[0].Key)
	}
	if items[0].Data.ItemType != "book" {
		t.Errorf("items[0].Data.ItemType = %v, want book", items[0].Data.ItemType)
	}

	// Test second item
	if items[1].Key != "EFGH5678" {
		t.Errorf("items[1].Key = %v, want EFGH5678", items[1].Key)
	}
	if items[1].Data.ItemType != "journalArticle" {
		t.Errorf("items[1].Data.ItemType = %v, want journalArticle", items[1].Data.ItemType)
	}
}

func TestCollectionUnmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/collections.json")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	var collections []Collection
	if err := json.Unmarshal(data, &collections); err != nil {
		t.Fatalf("failed to unmarshal collections: %v", err)
	}

	if len(collections) != 2 {
		t.Fatalf("len(collections) = %v, want 2", len(collections))
	}

	// Test first collection
	if collections[0].Key != "COLL1234" {
		t.Errorf("collections[0].Key = %v, want COLL1234", collections[0].Key)
	}
	if collections[0].Version != 50 {
		t.Errorf("collections[0].Version = %v, want 50", collections[0].Version)
	}
	if collections[0].Data.Name != "Test Collection" {
		t.Errorf("collections[0].Data.Name = %v, want Test Collection", collections[0].Data.Name)
	}
	if collections[0].Data.ParentCollection != "" {
		t.Errorf("collections[0].Data.ParentCollection = %v, want empty string", collections[0].Data.ParentCollection)
	}
	if collections[0].Meta.NumItems != 5 {
		t.Errorf("collections[0].Meta.NumItems = %v, want 5", collections[0].Meta.NumItems)
	}

	// Test second collection (subcollection)
	if collections[1].Key != "COLL5678" {
		t.Errorf("collections[1].Key = %v, want COLL5678", collections[1].Key)
	}
	if collections[1].Data.Name != "Subcollection" {
		t.Errorf("collections[1].Data.Name = %v, want Subcollection", collections[1].Data.Name)
	}
	if collections[1].Data.ParentCollection != "COLL1234" {
		t.Errorf("collections[1].Data.ParentCollection = %v, want COLL1234", collections[1].Data.ParentCollection)
	}
}

func TestGroupUnmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/groups.json")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	var groups []Group
	if err := json.Unmarshal(data, &groups); err != nil {
		t.Fatalf("failed to unmarshal groups: %v", err)
	}

	if len(groups) != 1 {
		t.Fatalf("len(groups) = %v, want 1", len(groups))
	}

	group := groups[0]
	if group.ID != 169947 {
		t.Errorf("ID = %v, want 169947", group.ID)
	}
	if group.Name != "smart_cities" {
		t.Errorf("Name = %v, want smart_cities", group.Name)
	}
	if group.Type != "Private" {
		t.Errorf("Type = %v, want Private", group.Type)
	}
	if group.Owner != 436 {
		t.Errorf("Owner = %v, want 436", group.Owner)
	}
	if len(group.Members) != 2 {
		t.Errorf("len(Members) = %v, want 2", len(group.Members))
	}
	if group.Meta.NumItems != 817 {
		t.Errorf("Meta.NumItems = %v, want 817", group.Meta.NumItems)
	}

	// Test timestamp parsing
	expectedCreated, _ := time.Parse(time.RFC3339, "2013-05-22T11:22:46Z")
	if !group.Meta.Created.Equal(expectedCreated) {
		t.Errorf("Meta.Created = %v, want %v", group.Meta.Created, expectedCreated)
	}
}

func TestTagsResponseUnmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/tags.json")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	var tags []TagsResponse
	if err := json.Unmarshal(data, &tags); err != nil {
		t.Fatalf("failed to unmarshal tags: %v", err)
	}

	if len(tags) != 2 {
		t.Fatalf("len(tags) = %v, want 2", len(tags))
	}

	// Test first tag
	if tags[0].Tag != "test" {
		t.Errorf("tags[0].Tag = %v, want test", tags[0].Tag)
	}
	if tags[0].Type != 0 {
		t.Errorf("tags[0].Type = %v, want 0", tags[0].Type)
	}
	if tags[0].NumItems != 5 {
		t.Errorf("tags[0].NumItems = %v, want 5", tags[0].NumItems)
	}

	// Test second tag
	if tags[1].Tag != "example" {
		t.Errorf("tags[1].Tag = %v, want example", tags[1].Tag)
	}
	if tags[1].Type != 1 {
		t.Errorf("tags[1].Type = %v, want 1", tags[1].Type)
	}
}

func TestCreator(t *testing.T) {
	// Test two-field creator
	twoField := Creator{
		CreatorType: "author",
		FirstName:   "John",
		LastName:    "Doe",
	}
	data, err := json.Marshal(twoField)
	if err != nil {
		t.Fatalf("failed to marshal creator: %v", err)
	}
	var unmarshaledTwoField Creator
	if err := json.Unmarshal(data, &unmarshaledTwoField); err != nil {
		t.Fatalf("failed to unmarshal creator: %v", err)
	}
	if unmarshaledTwoField.FirstName != "John" || unmarshaledTwoField.LastName != "Doe" {
		t.Errorf("two-field creator not unmarshaled correctly: %+v", unmarshaledTwoField)
	}

	// Test single-field creator
	singleField := Creator{
		CreatorType: "author",
		Name:        "John Doe",
	}
	data, err = json.Marshal(singleField)
	if err != nil {
		t.Fatalf("failed to marshal creator: %v", err)
	}
	var unmarshaledSingleField Creator
	if err := json.Unmarshal(data, &unmarshaledSingleField); err != nil {
		t.Fatalf("failed to unmarshal creator: %v", err)
	}
	if unmarshaledSingleField.Name != "John Doe" {
		t.Errorf("single-field creator not unmarshaled correctly: %+v", unmarshaledSingleField)
	}
}

func TestTag(t *testing.T) {
	// Test automatic tag
	autoTag := Tag{
		Tag:  "automatic",
		Type: 0,
	}
	data, err := json.Marshal(autoTag)
	if err != nil {
		t.Fatalf("failed to marshal tag: %v", err)
	}
	var unmarshaledAutoTag Tag
	if err := json.Unmarshal(data, &unmarshaledAutoTag); err != nil {
		t.Fatalf("failed to unmarshal tag: %v", err)
	}
	if unmarshaledAutoTag.Tag != "automatic" || unmarshaledAutoTag.Type != 0 {
		t.Errorf("automatic tag not unmarshaled correctly: %+v", unmarshaledAutoTag)
	}

	// Test manual tag
	manualTag := Tag{
		Tag:  "manual",
		Type: 1,
	}
	data, err = json.Marshal(manualTag)
	if err != nil {
		t.Fatalf("failed to marshal tag: %v", err)
	}
	var unmarshaledManualTag Tag
	if err := json.Unmarshal(data, &unmarshaledManualTag); err != nil {
		t.Fatalf("failed to unmarshal tag: %v", err)
	}
	if unmarshaledManualTag.Tag != "manual" || unmarshaledManualTag.Type != 1 {
		t.Errorf("manual tag not unmarshaled correctly: %+v", unmarshaledManualTag)
	}
}
