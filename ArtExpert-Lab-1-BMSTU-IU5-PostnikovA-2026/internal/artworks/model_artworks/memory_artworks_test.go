package model_artworks

import (
	"testing"
	"time"
)

func TestArtworkFeedStartsAtNobleman(t *testing.T) {
	artwork, ok := NewMemoryArtworkLookup().OpenArtworkFeed(0, false)
	if !ok || artwork.ArtworkID != 9 {
		t.Fatalf("feed start = %+v ok=%v", artwork, ok)
	}
}

func TestArtworkFeedNextIsFollowingArtwork(t *testing.T) {
	artwork, ok := NewMemoryArtworkLookup().OpenArtworkFeed(9, true)
	if !ok || artwork.ArtworkID != 10 {
		t.Fatalf("next artwork = %+v ok=%v", artwork, ok)
	}
}

func TestArtworkFeedHidesDraftAndRemoved(t *testing.T) {
	lookup := NewMemoryArtworkLookup()
	if _, ok := lookup.OpenArtworkFeed(21, false); ok {
		t.Fatal("draft opened in artwork feed")
	}
	if _, ok := lookup.OpenArtworkFeed(30, false); ok {
		t.Fatal("removed artwork opened in artwork feed")
	}
}

func TestCreatedOnKeepsNotLaterPublished(t *testing.T) {
	published := NewMemoryArtworkLookup().PublishedArtworks(artworkDate(1580, time.December, 31), true)
	if len(published) != 4 {
		t.Fatalf("len = %d", len(published))
	}
	for _, artwork := range published {
		limit := artworkDate(1580, time.December, 31)
		if !(artwork.CreatedOn.Unix() <= limit.Unix()) || artwork.ArtworkStatus != StatusArtworkPublished {
			t.Fatalf("unexpected %+v", artwork)
		}
	}
}

func TestRemovedAbsentFromCreatedOnTile(t *testing.T) {
	for _, artwork := range NewMemoryArtworkLookup().PublishedArtworks(time.Time{}, false) {
		if artwork.ArtworkTitle == "The Three Graces" || artwork.ArtworkStatus != StatusArtworkPublished {
			t.Fatalf("tile leaked %+v", artwork)
		}
	}
}

func TestOnlyArtworkDraft(t *testing.T) {
	artwork, ok := NewMemoryArtworkLookup().OnlyArtworkDraft()
	if !ok || artwork.ArtworkID != 21 || artwork.ArtworkStatus != StatusArtworkDraft {
		t.Fatalf("draft = %+v ok=%v", artwork, ok)
	}
}
