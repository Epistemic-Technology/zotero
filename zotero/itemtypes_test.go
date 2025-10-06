package zotero

import "testing"

func TestIsExcludeFilter(t *testing.T) {
	tests := []struct {
		name     string
		itemType string
		want     bool
	}{
		{
			name:     "exclude annotation",
			itemType: "-annotation",
			want:     true,
		},
		{
			name:     "exclude note",
			itemType: "-note",
			want:     true,
		},
		{
			name:     "include book",
			itemType: "book",
			want:     false,
		},
		{
			name:     "include journalArticle",
			itemType: "journalArticle",
			want:     false,
		},
		{
			name:     "empty string",
			itemType: "",
			want:     false,
		},
		{
			name:     "just hyphen",
			itemType: "-",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsExcludeFilter(tt.itemType); got != tt.want {
				t.Errorf("IsExcludeFilter(%q) = %v, want %v", tt.itemType, got, tt.want)
			}
		})
	}
}

func TestWithoutExcludePrefix(t *testing.T) {
	tests := []struct {
		name     string
		itemType string
		want     string
	}{
		{
			name:     "exclude annotation",
			itemType: "-annotation",
			want:     "annotation",
		},
		{
			name:     "exclude note",
			itemType: "-note",
			want:     "note",
		},
		{
			name:     "include book",
			itemType: "book",
			want:     "book",
		},
		{
			name:     "include journalArticle",
			itemType: "journalArticle",
			want:     "journalArticle",
		},
		{
			name:     "empty string",
			itemType: "",
			want:     "",
		},
		{
			name:     "just hyphen",
			itemType: "-",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithoutExcludePrefix(tt.itemType); got != tt.want {
				t.Errorf("WithoutExcludePrefix(%q) = %v, want %v", tt.itemType, got, tt.want)
			}
		})
	}
}

func TestItemTypeConstants(t *testing.T) {
	// Verify that some key constants have the expected values
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"book constant", ItemTypeBook, "book"},
		{"journalArticle constant", ItemTypeJournalArticle, "journalArticle"},
		{"webpage constant", ItemTypeWebpage, "webpage"},
		{"annotation constant", ItemTypeAnnotation, "annotation"},
		{"note constant", ItemTypeNote, "note"},
		{"attachment constant", ItemTypeAttachment, "attachment"},
		{"conferencePaper constant", ItemTypeConferencePaper, "conferencePaper"},
		{"thesis constant", ItemTypeThesis, "thesis"},
		{"podcast constant", ItemTypePodcast, "podcast"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.want)
			}
		})
	}
}

func TestCreatorTypeConstants(t *testing.T) {
	// Verify that creator type constants have the expected values
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"author constant", CreatorTypeAuthor, "author"},
		{"editor constant", CreatorTypeEditor, "editor"},
		{"contributor constant", CreatorTypeContributor, "contributor"},
		{"translator constant", CreatorTypeTranslator, "translator"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.want)
			}
		})
	}
}
