package store_artworkdb

import (
	"errors"
	"strings"
	"time"

	"github.com/Anatole-Code/art-expert-backend/internal/artworks/model_artworkdb"
	"github.com/Anatole-Code/art-expert-backend/internal/artworks/model_artworks"
	"gorm.io/gorm"
)

const (
	unavailableArtworkImageURL = "http://127.0.0.1:9/missing-artwork-image"
	unavailableArtworkVideoURL = "http://127.0.0.1:9/missing-artwork-video"
)

func (lookup *ArtworkLookup) CreateArtworkDraft(title string) (model_artworks.ArtworkAttribution, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return model_artworks.ArtworkAttribution{}, model_artworks.ErrArtworkTitleMissing
	}
	var created model_artworkdb.Artwork
	err := lookup.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		err := tx.Model(&model_artworkdb.Artwork{}).
			Where("connoisseur_id = ? AND artwork_status = ?", model_artworks.CreatorConnoisseurID(), model_artworks.StatusArtworkDraft).
			Count(&count).Error
		if err != nil {
			return err
		}
		if count > 0 {
			return model_artworks.ErrArtworkDraftExists
		}
		imageURL := unavailableArtworkImageURL
		videoURL := unavailableArtworkVideoURL
		created = model_artworkdb.Artwork{
			ArtworkTitle:    title,
			ArtworkStatus:   model_artworks.StatusArtworkDraft,
			ArtworkImageURL: &imageURL,
			ArtworkVideoURL: &videoURL,
			RecordedAt:      time.Now().UTC(),
			ConnoisseurID:   int32(model_artworks.CreatorConnoisseurID()),
		}
		return tx.Create(&created).Error
	})
	if err != nil {
		return model_artworks.ArtworkAttribution{}, err
	}
	artworks, err := lookup.attributions([]model_artworkdb.Artwork{created})
	if err != nil || len(artworks) != 1 {
		return model_artworks.ArtworkAttribution{}, err
	}
	return artworks[0], nil
}

func (lookup *ArtworkLookup) PublishArtworkDraft(note string, createdOn time.Time, diagonal int) (model_artworks.ArtworkAttribution, error) {
	note = strings.TrimSpace(note)
	if note == "" || createdOn.IsZero() || diagonal <= 0 {
		return model_artworks.ArtworkAttribution{}, model_artworks.ErrArtworkPublicationIncomplete
	}
	var saved model_artworkdb.Artwork
	err := lookup.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("connoisseur_id = ? AND artwork_status = ?", model_artworks.CreatorConnoisseurID(), model_artworks.StatusArtworkDraft).
			Take(&saved).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model_artworks.ErrArtworkDraftMissing
		}
		if err != nil {
			return err
		}
		diagonalCm := int32(diagonal)
		formedAt := time.Now().UTC()
		saved.AttributionNote = &note
		saved.CreatedOn = &createdOn
		saved.DiagonalCm = &diagonalCm
		saved.ArtworkStatus = model_artworks.StatusArtworkPublished
		saved.NoteFormedAt = &formedAt
		return tx.Save(&saved).Error
	})
	if err != nil {
		return model_artworks.ArtworkAttribution{}, err
	}
	artworks, err := lookup.attributions([]model_artworkdb.Artwork{saved})
	if err != nil || len(artworks) != 1 {
		return model_artworks.ArtworkAttribution{}, err
	}
	return artworks[0], nil
}

func (lookup *ArtworkLookup) RemovePublishedArtwork(id int) error {
	result := lookup.db.Exec(
		`UPDATE artworks SET artwork_status = ? WHERE artwork_id = ? AND artwork_status = ?`,
		model_artworks.StatusArtworkRemoved,
		id,
		model_artworks.StatusArtworkPublished,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return model_artworks.ErrArtworkNotPublished
	}
	return nil
}
