package model_artworks

import "time"

type ArtworkLookup interface {
	PublishedArtworks(createdOn time.Time, createdOnEnabled bool) []ArtworkAttribution
	OpenArtworkFeed(id int, nextArtwork bool) (ArtworkAttribution, bool)
	OnlyArtworkDraft() (ArtworkAttribution, bool)
	CreateArtworkDraft(title string) (ArtworkAttribution, error)
	PublishArtworkDraft(note string, createdOn time.Time, diagonal int) (ArtworkAttribution, error)
	RemovePublishedArtwork(id int) error
}
