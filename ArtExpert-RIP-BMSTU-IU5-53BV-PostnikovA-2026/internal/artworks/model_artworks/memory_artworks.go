package model_artworks

import (
	"strings"
	"time"
)

type MemoryArtworkLookup struct {
	storedArtworks []ArtworkAttribution
}

func NewMemoryArtworkLookup() *MemoryArtworkLookup {
	at := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	return &MemoryArtworkLookup{storedArtworks: []ArtworkAttribution{
		{
			ArtworkID: 9, ArtworkTitle: "The Nobleman with his Hand on his Chest",
			AttributionNote: "Портрет кавалера с рукой на груди. Эль Греко, холст, масло.",
			ArtworkStatus:   StatusArtworkPublished, ArtworkImageKey: "nobleman-hand-chest.jpg", ArtworkVideoKey: "nobleman-hand-chest.mp4",
			CreatedOn: artworkDate(1580, 6, 1), DiagonalCm: 81, PainterName: "El Greco",
			RecordedAt: at, ConnoisseurID: 1, NoteFormedAt: at, MarkConnoisseurIDs: []int{1, 2, 3},
		},
		{
			ArtworkID: 10, ArtworkTitle: "The Virgin Mary",
			AttributionNote: "Богоматерь. Эль Греко, холст, масло.",
			ArtworkStatus:   StatusArtworkPublished, ArtworkImageKey: "virgin-mary.jpg", ArtworkVideoKey: "virgin-mary.mp4",
			CreatedOn: artworkDate(1590, 3, 15), DiagonalCm: 52, PainterName: "El Greco",
			RecordedAt: at, ConnoisseurID: 1, NoteFormedAt: at, MarkConnoisseurIDs: []int{2},
		},
		{
			ArtworkID: 11, ArtworkTitle: "Christ Carrying the Cross",
			AttributionNote: "Несение креста. Эль Греко, холст, масло.",
			ArtworkStatus:   StatusArtworkPublished, ArtworkImageKey: "christ-carrying-cross.jpg", ArtworkVideoKey: "christ-carrying-cross.mp4",
			CreatedOn: artworkDate(1580, 11, 2), DiagonalCm: 66, PainterName: "El Greco",
			RecordedAt: at, ConnoisseurID: 1, NoteFormedAt: at, MarkConnoisseurIDs: []int{1, 4},
		},
		{
			ArtworkID: 14, ArtworkTitle: "David with the Head of Goliath",
			AttributionNote: "Давид с головой Голиафа. Караваджо, холст, масло.",
			ArtworkStatus:   StatusArtworkPublished, ArtworkImageKey: "david-goliath.jpg", ArtworkVideoKey: "david-goliath.mp4",
			CreatedOn: artworkDate(1610, 4, 8), DiagonalCm: 91, PainterName: "Caravaggio",
			RecordedAt: at, ConnoisseurID: 1, NoteFormedAt: at, MarkConnoisseurIDs: []int{1},
		},
		{
			ArtworkID: 15, ArtworkTitle: "Self-portrait",
			AttributionNote: "Автопортрет. Тициан, холст, масло.",
			ArtworkStatus:   StatusArtworkPublished, ArtworkImageKey: "self-portrait-titian.jpg", ArtworkVideoKey: "self-portrait-titian.mp4",
			CreatedOn: artworkDate(1562, 8, 20), DiagonalCm: 86, PainterName: "Titian",
			RecordedAt: at, ConnoisseurID: 1, NoteFormedAt: at, MarkConnoisseurIDs: []int{3, 5},
		},
		{
			ArtworkID: 16, ArtworkTitle: "The Cardinal",
			AttributionNote: "Портрет кардинала. Рафаэль, дерево, масло.",
			ArtworkStatus:   StatusArtworkPublished, ArtworkImageKey: "the-cardinal.jpg", ArtworkVideoKey: "the-cardinal.mp4",
			CreatedOn: artworkDate(1510, 1, 12), DiagonalCm: 79, PainterName: "Raphael",
			RecordedAt: at, ConnoisseurID: 1, NoteFormedAt: at, MarkConnoisseurIDs: []int{2, 5},
		},
		{
			ArtworkID: 21, ArtworkTitle: "The Virgin of the Chair",
			AttributionNote: "Гвидо Рени",
			ArtworkStatus:   StatusArtworkDraft, ArtworkImageKey: "virgin-of-the-chair.jpg", ArtworkVideoKey: "virgin-of-the-chair.mp4",
			CreatedOn: artworkDate(1625, 5, 1), DiagonalCm: 68, PainterName: "Guido Reni",
			RecordedAt: at, ConnoisseurID: 1, NoteFormedAt: at, MarkConnoisseurIDs: []int{},
		},
		{
			ArtworkID: 30, ArtworkTitle: "The Three Graces",
			AttributionNote: "Три грации. Рубенс. Карточка снята.",
			ArtworkStatus:   StatusArtworkRemoved, ArtworkImageKey: "three-graces.jpg", ArtworkVideoKey: "three-graces.mp4",
			CreatedOn: artworkDate(1635, 7, 1), DiagonalCm: 181, PainterName: "Peter Paul Rubens",
			RecordedAt: at, ConnoisseurID: 1, NoteFormedAt: at, MarkConnoisseurIDs: []int{1},
		},
	}}
}

func (lookup *MemoryArtworkLookup) PublishedArtworks(createdOn time.Time, createdOnEnabled bool) []ArtworkAttribution {
	limit := artworkDate(createdOn.Date())
	published := make([]ArtworkAttribution, 0)
	for _, artwork := range lookup.storedArtworks {
		if artwork.ArtworkStatus != StatusArtworkPublished {
			continue
		}
		created := artworkDate(artwork.CreatedOn.Date())
		if createdOnEnabled && !(created.Unix() <= limit.Unix()) {
			continue
		}
		published = append(published, artwork)
	}
	return published
}

func artworkDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func (lookup *MemoryArtworkLookup) OpenArtworkFeed(id int, nextArtwork bool) (ArtworkAttribution, bool) {
	published := lookup.PublishedArtworks(time.Time{}, false)
	if len(published) == 0 {
		return ArtworkAttribution{}, false
	}
	if id == 0 {
		return published[0], true
	}
	index := -1
	for i, artwork := range published {
		if artwork.ArtworkID == id {
			index = i
			break
		}
	}
	if index < 0 {
		return ArtworkAttribution{}, false
	}
	if !nextArtwork {
		return published[index], true
	}
	return published[(index+1)%len(published)], true
}

func (lookup *MemoryArtworkLookup) OnlyArtworkDraft() (ArtworkAttribution, bool) {
	var found ArtworkAttribution
	seen := 0
	for _, artwork := range lookup.storedArtworks {
		if artwork.ArtworkStatus != StatusArtworkDraft {
			continue
		}
		seen++
		found = artwork
	}
	if seen != 1 {
		return ArtworkAttribution{}, false
	}
	return found, true
}

func (lookup *MemoryArtworkLookup) CreateArtworkDraft(title string) (ArtworkAttribution, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return ArtworkAttribution{}, ErrArtworkTitleMissing
	}
	if _, ok := lookup.OnlyArtworkDraft(); ok {
		return ArtworkAttribution{}, ErrArtworkDraftExists
	}
	nextID := 1
	for _, artwork := range lookup.storedArtworks {
		if artwork.ArtworkID >= nextID {
			nextID = artwork.ArtworkID + 1
		}
	}
	draft := ArtworkAttribution{
		ArtworkID:       nextID,
		ArtworkTitle:    title,
		ArtworkStatus:   StatusArtworkDraft,
		ArtworkImageKey: "missing-artwork-image",
		ArtworkVideoKey: "missing-artwork-video",
		RecordedAt:      time.Now().UTC(),
		ConnoisseurID:   CreatorConnoisseurID(),
	}
	lookup.storedArtworks = append(lookup.storedArtworks, draft)
	return draft, nil
}

func (lookup *MemoryArtworkLookup) PublishArtworkDraft(note string, createdOn time.Time, diagonal int) (ArtworkAttribution, error) {
	note = strings.TrimSpace(note)
	if note == "" || createdOn.IsZero() || diagonal <= 0 {
		return ArtworkAttribution{}, ErrArtworkPublicationIncomplete
	}
	for i, artwork := range lookup.storedArtworks {
		if artwork.ArtworkStatus != StatusArtworkDraft || artwork.ConnoisseurID != CreatorConnoisseurID() {
			continue
		}
		artwork.AttributionNote = note
		artwork.CreatedOn = createdOn
		artwork.DiagonalCm = diagonal
		artwork.ArtworkStatus = StatusArtworkPublished
		artwork.NoteFormedAt = time.Now().UTC()
		lookup.storedArtworks[i] = artwork
		return artwork, nil
	}
	return ArtworkAttribution{}, ErrArtworkDraftMissing
}

func (lookup *MemoryArtworkLookup) RemovePublishedArtwork(id int) error {
	for i, artwork := range lookup.storedArtworks {
		if artwork.ArtworkID == id && artwork.ArtworkStatus == StatusArtworkPublished {
			artwork.ArtworkStatus = StatusArtworkRemoved
			lookup.storedArtworks[i] = artwork
			return nil
		}
	}
	return ErrArtworkNotPublished
}
