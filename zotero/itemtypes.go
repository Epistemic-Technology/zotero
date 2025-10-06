package zotero

// Common Zotero item types as string constants.
// These provide IDE autocomplete and type safety for the most frequently used item types.
// You can still use raw strings for any item type not listed here.
//
// For a complete list of all item types, use the ItemTypes() method to fetch the current
// schema from the Zotero API, or see: https://api.zotero.org/schema
const (
	// Regular item types
	ItemTypeBook             = "book"
	ItemTypeBookSection      = "bookSection"
	ItemTypeJournalArticle   = "journalArticle"
	ItemTypeMagazineArticle  = "magazineArticle"
	ItemTypeNewspaperArticle = "newspaperArticle"
	ItemTypeConferencePaper  = "conferencePaper"
	ItemTypeThesis           = "thesis"
	ItemTypeReport           = "report"
	ItemTypeWebpage          = "webpage"
	ItemTypeBlogPost         = "blogPost"
	ItemTypeForumPost        = "forumPost"
	ItemTypePreprint         = "preprint"
	ItemTypeManuscript       = "manuscript"
	ItemTypePresentation     = "presentation"

	// Media types
	ItemTypePodcast        = "podcast"
	ItemTypeVideoRecording = "videoRecording"
	ItemTypeAudioRecording = "audioRecording"
	ItemTypeFilm           = "film"

	// Legal/Government
	ItemTypeCase    = "case"
	ItemTypeStatute = "statute"
	ItemTypeBill    = "bill"
	ItemTypePatent  = "patent"
	ItemTypeHearing = "hearing"

	// Reference types
	ItemTypeDictionaryEntry     = "dictionaryEntry"
	ItemTypeEncyclopediaArticle = "encyclopediaArticle"

	// Other media
	ItemTypeArtwork         = "artwork"
	ItemTypeMap             = "map"
	ItemTypeEmail           = "email"
	ItemTypeLetter          = "letter"
	ItemTypeInterview       = "interview"
	ItemTypeInstantMessage  = "instantMessage"
	ItemTypeDocument        = "document"
	ItemTypeComputerProgram = "computerProgram"
	ItemTypeDataset         = "dataset"
	ItemTypeStandard        = "standard"
	ItemTypeTVBroadcast     = "tvBroadcast"
	ItemTypeRadioBroadcast  = "radioBroadcast"

	// Special item types
	ItemTypeAttachment = "attachment"
	ItemTypeNote       = "note"
	ItemTypeAnnotation = "annotation"
)

// Common creator types as string constants.
// Creator types vary by item type; use the CreatorTypes() method for item-specific creator types.
const (
	CreatorTypeAuthor       = "author"
	CreatorTypeEditor       = "editor"
	CreatorTypeContributor  = "contributor"
	CreatorTypeTranslator   = "translator"
	CreatorTypeSeriesEditor = "seriesEditor"
	CreatorTypeReviewer     = "reviewer"
)

// IsExcludeFilter returns true if the item type string represents an exclusion filter.
// Exclusion filters are prefixed with "-" (e.g., "-annotation" excludes annotations).
func IsExcludeFilter(itemType string) bool {
	return len(itemType) > 0 && itemType[0] == '-'
}

// WithoutExcludePrefix returns the item type string with any leading "-" removed.
func WithoutExcludePrefix(itemType string) string {
	if IsExcludeFilter(itemType) {
		return itemType[1:]
	}
	return itemType
}
