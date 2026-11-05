package model_artworks

import "time"

const (
	StatusArtworkDraft     = "artwork_draft"
	StatusArtworkPublished = "artwork_published"
	StatusArtworkRemoved   = "artwork_removed"
)

type ArtworkAttribution struct {
	ArtworkID         int
	ArtworkTitle      string
	AttributionNote    string
	ArtworkStatus     string
	ArtworkImageKey   string
	ArtworkVideoKey   string
	CreatedOn          time.Time
	DiagonalCm         int
	PainterName        string
	RecordedAt         time.Time
	ConnoisseurID      int
	NoteFormedAt       time.Time
	MarkConnoisseurIDs []int
}
