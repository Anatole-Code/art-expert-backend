package store_artworkdb

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Anatole-Code/art-expert-backend/internal/artworks/config_artworkserver"
	"github.com/Anatole-Code/art-expert-backend/internal/artworks/dsn_artworkdb"
	"github.com/Anatole-Code/art-expert-backend/internal/artworks/model_artworks"
	"github.com/joho/godotenv"
)

func TestArtworkLookupReadsDatabase(t *testing.T) {
	root, err := config_artworkserver.FindProjectRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := godotenv.Load(filepath.Join(root, ".env")); err != nil {
		t.Fatal(err)
	}
	db, err := Open(dsn_artworkdb.FromEnv())
	if err != nil {
		t.Fatal(err)
	}
	lookup := NewArtworkLookup(db)

	artwork, ok := lookup.OpenArtworkFeed(0, false)
	if !ok || artwork.ArtworkID != 9 || artwork.ArtworkTitle != "The Nobleman with his Hand on his Chest" || len(artwork.MarkConnoisseurIDs) != 3 {
		t.Fatalf("feed start = %+v ok=%v", artwork, ok)
	}
	if artwork.ArtworkImageKey != "nobleman-hand-chest.jpg" || artwork.ArtworkVideoKey != "nobleman-hand-chest.mp4" {
		t.Fatalf("storage keys = %s %s", artwork.ArtworkImageKey, artwork.ArtworkVideoKey)
	}

	next, ok := lookup.OpenArtworkFeed(9, true)
	if !ok || next.ArtworkID != 10 {
		t.Fatalf("next artwork = %+v ok=%v", next, ok)
	}
	wrapped, ok := lookup.OpenArtworkFeed(16, true)
	if !ok || wrapped.ArtworkID != 9 {
		t.Fatalf("wrapped artwork = %+v ok=%v", wrapped, ok)
	}
	if _, ok := lookup.OpenArtworkFeed(21, false); ok {
		t.Fatal("draft opened in artwork feed")
	}
	if _, ok := lookup.OpenArtworkFeed(30, true); ok {
		t.Fatal("removed artwork opened in artwork feed")
	}

	published := lookup.PublishedArtworks(time.Date(1580, time.December, 31, 0, 0, 0, 0, time.UTC), true)
	if len(published) != 4 {
		t.Fatalf("published len = %d", len(published))
	}
	for _, item := range published {
		if item.ArtworkStatus != model_artworks.StatusArtworkPublished || item.CreatedOn.After(time.Date(1580, time.December, 31, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected %+v", item)
		}
	}

	draft, ok := lookup.OnlyArtworkDraft()
	if !ok || draft.ArtworkID != 21 || draft.ArtworkStatus != model_artworks.StatusArtworkDraft {
		t.Fatalf("draft = %+v ok=%v", draft, ok)
	}
}

func TestArtworkCommandsRollback(t *testing.T) {
	root, err := config_artworkserver.FindProjectRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := godotenv.Load(filepath.Join(root, ".env")); err != nil {
		t.Fatal(err)
	}
	db, err := Open(dsn_artworkdb.FromEnv())
	if err != nil {
		t.Fatal(err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	lookup := NewArtworkLookup(tx)
	if _, err := lookup.CreateArtworkDraft("Лишняя"); !errors.Is(err, model_artworks.ErrArtworkDraftExists) {
		t.Fatal(err)
	}
	published, err := lookup.PublishArtworkDraft("Проверка публикации", time.Date(1580, time.June, 1, 0, 0, 0, 0, time.UTC), 40)
	if err != nil || published.ArtworkID != 21 || published.ArtworkStatus != model_artworks.StatusArtworkPublished || published.NoteFormedAt.IsZero() || published.DiagonalCm != 40 {
		t.Fatalf("publish %+v %v", published, err)
	}
	draft, err := lookup.CreateArtworkDraft("Новая картина")
	if err != nil {
		t.Fatal(err)
	}
	var imageURL, videoURL string
	if err := tx.Raw("SELECT artwork_image_url, artwork_video_url FROM artworks WHERE artwork_id = ?", draft.ArtworkID).Row().Scan(&imageURL, &videoURL); err != nil {
		t.Fatal(err)
	}
	if imageURL != unavailableArtworkImageURL || videoURL != unavailableArtworkVideoURL || draft.ConnoisseurID != model_artworks.CreatorConnoisseurID() {
		t.Fatalf("urls %s %s draft %+v", imageURL, videoURL, draft)
	}
	if _, err := lookup.CreateArtworkDraft("Ещё одна"); !errors.Is(err, model_artworks.ErrArtworkDraftExists) {
		t.Fatal(err)
	}
	if err := lookup.RemovePublishedArtwork(draft.ArtworkID); !errors.Is(err, model_artworks.ErrArtworkNotPublished) {
		t.Fatal(err)
	}
	publishedDraft, err := lookup.PublishArtworkDraft("Описание новой", time.Date(1600, time.January, 2, 0, 0, 0, 0, time.UTC), 33)
	if err != nil || publishedDraft.ArtworkID != draft.ArtworkID {
		t.Fatalf("new publish %+v %v", publishedDraft, err)
	}
	if err := lookup.RemovePublishedArtwork(publishedDraft.ArtworkID); err != nil {
		t.Fatal(err)
	}
	if _, ok := lookup.OpenArtworkFeed(publishedDraft.ArtworkID, false); ok {
		t.Fatal("removed artwork opened")
	}
	if err := tx.Rollback().Error; err != nil {
		t.Fatal(err)
	}
	var status string
	var count int
	if err := db.Raw("SELECT artwork_status FROM artworks WHERE artwork_id = 21").Scan(&status).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Raw("SELECT COUNT(*) FROM artworks").Scan(&count).Error; err != nil {
		t.Fatal(err)
	}
	if status != model_artworks.StatusArtworkDraft || count != 8 {
		t.Fatalf("database changed status=%s count=%d", status, count)
	}
}
