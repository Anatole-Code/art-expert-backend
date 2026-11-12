package store_artworkdb

import (
	"path"
	"strings"
	"time"

	"github.com/Anatole-Code/art-expert-backend/internal/artworks/model_artworkdb"
	"github.com/Anatole-Code/art-expert-backend/internal/artworks/model_artworks"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ArtworkLookup struct {
	db *gorm.DB
}

func NewArtworkLookup(db *gorm.DB) *ArtworkLookup {
	return &ArtworkLookup{db: db}
}

func (lookup *ArtworkLookup) PublishedArtworks(createdOn time.Time, createdOnEnabled bool) []model_artworks.ArtworkAttribution {
	query := lookup.db.Model(&model_artworkdb.Artwork{}).
		Where("artwork_status = ?", model_artworks.StatusArtworkPublished)
	if createdOnEnabled {
		limit := time.Date(createdOn.Year(), createdOn.Month(), createdOn.Day(), 0, 0, 0, 0, time.UTC)
		query = query.Where("created_on <= ?", limit)
	}
	var rows []model_artworkdb.Artwork
	if err := query.Order("artwork_id").Find(&rows).Error; err != nil {
		return nil
	}
	artworks, err := lookup.attributions(rows)
	if err != nil {
		return nil
	}
	return artworks
}

func (lookup *ArtworkLookup) OpenArtworkFeed(id int, nextArtwork bool) (model_artworks.ArtworkAttribution, bool) {
	query := lookup.db.Model(&model_artworkdb.Artwork{}).
		Where("artwork_status = ?", model_artworks.StatusArtworkPublished)
	var err error
	var row model_artworkdb.Artwork
	switch {
	case id == 0:
		err = query.Order("artwork_id").Limit(1).Take(&row).Error
	case nextArtwork:
		err = query.
			Where(
				"EXISTS (SELECT 1 FROM artworks AS current_artwork WHERE current_artwork.artwork_id = ? AND current_artwork.artwork_status = ?)",
				id,
				model_artworks.StatusArtworkPublished,
			).
			Clauses(clause.OrderBy{Expression: clause.Expr{SQL: "CASE WHEN artwork_id > ? THEN 0 ELSE 1 END, artwork_id", Vars: []any{id}}}).
			Limit(1).
			Take(&row).Error
	default:
		err = query.Where("artwork_id = ?", id).Take(&row).Error
	}
	if err != nil {
		return model_artworks.ArtworkAttribution{}, false
	}
	artworks, err := lookup.attributions([]model_artworkdb.Artwork{row})
	if err != nil || len(artworks) != 1 {
		return model_artworks.ArtworkAttribution{}, false
	}
	return artworks[0], true
}

func (lookup *ArtworkLookup) OnlyArtworkDraft() (model_artworks.ArtworkAttribution, bool) {
	var rows []model_artworkdb.Artwork
	if err := lookup.db.Where("artwork_status = ?", model_artworks.StatusArtworkDraft).Find(&rows).Error; err != nil {
		return model_artworks.ArtworkAttribution{}, false
	}
	if len(rows) != 1 {
		return model_artworks.ArtworkAttribution{}, false
	}
	artworks, err := lookup.attributions(rows)
	if err != nil || len(artworks) != 1 {
		return model_artworks.ArtworkAttribution{}, false
	}
	return artworks[0], true
}

func (lookup *ArtworkLookup) attributions(rows []model_artworkdb.Artwork) ([]model_artworks.ArtworkAttribution, error) {
	ids := make([]int32, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ArtworkID)
	}
	marks, err := lookup.markIDs(ids)
	if err != nil {
		return nil, err
	}
	artworks := make([]model_artworks.ArtworkAttribution, 0, len(rows))
	for _, row := range rows {
		noteFormedAt := time.Time{}
		if row.NoteFormedAt != nil {
			noteFormedAt = *row.NoteFormedAt
		}
		diagonal := 0
		if row.DiagonalCm != nil {
			diagonal = int(*row.DiagonalCm)
		}
		createdOn := time.Time{}
		if row.CreatedOn != nil {
			createdOn = *row.CreatedOn
		}
		artworks = append(artworks, model_artworks.ArtworkAttribution{
			ArtworkID:          int(row.ArtworkID),
			ArtworkTitle:       row.ArtworkTitle,
			AttributionNote:    textOf(row.AttributionNote),
			ArtworkStatus:      row.ArtworkStatus,
			ArtworkImageKey:    artworkStorageKey(row.ArtworkImageURL),
			ArtworkVideoKey:    artworkStorageKey(row.ArtworkVideoURL),
			CreatedOn:          createdOn,
			DiagonalCm:         diagonal,
			PainterName:        textOf(row.PainterName),
			RecordedAt:         row.RecordedAt,
			ConnoisseurID:      int(row.ConnoisseurID),
			NoteFormedAt:       noteFormedAt,
			MarkConnoisseurIDs: marks[row.ArtworkID],
		})
	}
	return artworks, nil
}

func (lookup *ArtworkLookup) markIDs(artworkIDs []int32) (map[int32][]int, error) {
	grouped := make(map[int32][]int, len(artworkIDs))
	if len(artworkIDs) == 0 {
		return grouped, nil
	}
	var rows []model_artworkdb.ArtworkMark
	err := lookup.db.Where("artwork_id IN ?", artworkIDs).Order("artwork_id, connoisseur_id").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		grouped[row.ArtworkID] = append(grouped[row.ArtworkID], int(row.ConnoisseurID))
	}
	return grouped, nil
}

func textOf(raw *string) string {
	if raw == nil {
		return ""
	}
	return *raw
}

func artworkStorageKey(raw *string) string {
	if raw == nil {
		return ""
	}
	key := strings.TrimSpace(*raw)
	if key == "" {
		return ""
	}
	if cut, _, found := strings.Cut(key, "?"); found {
		key = cut
	}
	return path.Base(key)
}
